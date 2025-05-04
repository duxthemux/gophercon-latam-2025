package api

import (
	"context"
	_ "embed"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"gophercon-2025/cmd/api/cache"
	"gophercon-2025/cmd/api/llm"
	"gophercon-2025/cmd/api/rag"
	"gophercon-2025/cmd/api/telemetry"
)

type Service struct {
	llm   *llm.Service
	rag   *rag.Service
	cache *cache.Service
	model string

	metricResponseTime metric.Float64Counter
}

type Option func(*Service)

func WithModel(model string) Option {
	return func(service *Service) {
		service.model = model
	}
}

func WithRag(rag *rag.Service) Option {
	return func(service *Service) {
		service.rag = rag
	}
}

func WithCache(cache *cache.Service) Option {
	return func(service *Service) {
		service.cache = cache
	}
}

func WithLlm(l *llm.Service) Option {
	return func(service *Service) {
		service.llm = l
	}
}

func (a *Service) middlewareTrace(hctx huma.Context, next func(huma.Context)) {
	opid := hctx.Operation().OperationID

	ctx, span := telemetry.Tracer.Start(hctx.Context(), opid)
	hctx = huma.WithContext(hctx, ctx)

	start := time.Now()
	defer func() {
		dur := time.Since(start)
		a.metricResponseTime.Add(ctx, float64(dur.Seconds()), metric.WithAttributes(attribute.String("op_id", opid)))
		span.End()
		slog.Info("Route served", "op", opid, "duration", dur.String())
	}()

	next(hctx)
}

func (a *Service) test(_ context.Context, _ *struct{}) (*struct{ Body string }, error) {
	return &struct{ Body string }{Body: "Hello - I am working"}, nil
}

//go:embed index.html
var indexHtml []byte

func New(opt ...Option) http.Handler {
	service := &Service{}

	for _, opt := range opt {
		opt(service)
	}

	mux := http.NewServeMux()

	humaApi := humago.New(mux, huma.DefaultConfig("gophercon-2025", "1.0.0"))
	humaApi.UseMiddleware(service.middlewareTrace)

	huma.Register(humaApi, huma.Operation{
		OperationID: "apiV1TestGet",
		Method:      "GET",
		Path:        "/api/v1/test",
		Description: "Checks API General Availability",
	}, service.test)

	service.setupApiRag(humaApi)
	service.setupApiStatus(humaApi)
	service.setupApiLlm(humaApi)
	service.setupApiCache(humaApi)

	var err error

	service.metricResponseTime, err = telemetry.Meter.Float64Counter("response_time_sec")
	if err != nil {
		panic(err)
	}

	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write(indexHtml) //nolint:errcheck
	})

	return mux
}
