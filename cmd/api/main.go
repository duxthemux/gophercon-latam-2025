package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	ollama_api "github.com/ollama/ollama/api"
	"github.com/philippgille/chromem-go"
	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/otel"

	"gophercon-2025/cmd/api/api"
	"gophercon-2025/cmd/api/cache"
	"gophercon-2025/cmd/api/llm"
	"gophercon-2025/cmd/api/rag"
	"gophercon-2025/cmd/api/telemetry"
	"gophercon-2025/cmd/api/tokenizer"
	"gophercon-2025/cmd/api/tool"
	"gophercon-2025/pkg/env"
)

func run(ctx context.Context, f *flags) (err error) {
	otelShutdown, err := telemetry.Setup(ctx, f.otelEp, f.SlogLevel())
	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	slog.SetDefault(telemetry.Logger)

	_, span := telemetry.Tracer.Start(ctx, "startup")

	vecDb, err := chromem.NewPersistentDB(f.vecDbPath, true)
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("postgres", f.toolDb)
	if err != nil {
		return err
	}
	defer db.Close()

	rs, err := db.Query("select version()")
	if err != nil {
		return err
	}
	defer rs.Close()

	if rs.Err() != nil {
		return err
	}

	if !rs.Next() {
		return errors.New("could not retrieve db version")
	}

	var dbVersion string
	if err = rs.Scan(&dbVersion); err != nil {
		return err
	}

	slog.Info("Using tool db", "version", dbVersion)

	toolSvc := tool.New(
		tool.WithDb(db),
		tool.WithTracer(telemetry.Tracer),
	)

	l, err := net.Listen("tcp", f.listeningAddr)
	if err != nil {
		return err
	}

	ragService, err := rag.New(
		rag.WithDb(vecDb),
		rag.WithEmbModel(f.embModel),
		rag.WithTracer(telemetry.Tracer),
		rag.WithOllamaEndpoint(f.llmEp),
	)
	if err != nil {
		return err
	}

	cacheService, err := cache.New(
		cache.WithDb(vecDb),
		cache.WithEmbModel(f.embModel),
		cache.WithTracer(telemetry.Tracer),
		cache.WithOllamaEndpoint(f.llmEp),
	)

	tokenizerService := tokenizer.New(
		tokenizer.WithPretrainedFromCache(f.tokenizerModel, f.tokenizerCache),
	)

	llmUrl, err := url.Parse(f.llmEp)
	if err != nil {
		return err
	}

	client := ollama_api.NewClient(llmUrl, http.DefaultClient)

	ver, err := client.Version(ctx)
	if err != nil {
		return err
	}

	slog.Info("Llm Connected", "ver", ver)

	llmService := llm.New(
		llm.WithCache(cacheService),
		llm.WithLlmModel(f.llmModel),
		llm.WithLogger(telemetry.Logger),
		llm.WithMinConfidenceRag(f.minConfidenceRag),
		llm.WithMinConfidenceTool(f.minConfidenceTool),
		llm.WithMinConfidenceCache(f.minConfidenceCache),
		llm.WithOllama(client),
		llm.WithRag(ragService),
		llm.WithTemperature(f.temperature),
		llm.WithTokenizer(tokenizerService),
		llm.WithTracer(telemetry.Tracer),
		llm.WithMetrics(telemetry.Meter),
		llm.WithTool(toolSvc),
	)

	mux := api.New(
		api.WithLlm(llmService),
		api.WithModel(f.llmModel),
		api.WithRag(ragService),
		api.WithCache(cacheService),
	)

	server := &http.Server{
		Handler:      mux,
		ReadTimeout:  10 * time.Minute,
		WriteTimeout: 10 * time.Minute,
		IdleTimeout:  10 * time.Minute,
	}

	go func() {
		<-ctx.Done()

		if closeErr := server.Close(); closeErr != nil {
			slog.Warn("failed to close http server", "err", closeErr)
		}
	}()

	slog.Info("starting server")

	metricUp, err := otel.Meter("llm-api").Int64Gauge("up")
	if err != nil {
		return err
	}

	metricUp.Record(ctx, 1)

	defer func() {
		metricUp.Record(ctx, 0)
	}()

	span.End()

	err = server.Serve(l)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	env.Load("api")

	f := &flags{}

	if err := (&cli.Command{
		Name:  "api",
		Usage: "Start the api server",
		Flags: f.build(),
		Action: func(ctx context.Context, command *cli.Command) error {
			return run(ctx, f)
		},
	}).Run(ctx, os.Args); err != nil {
		panic(err)
	}

	slog.Info("shutdown complete")
}
