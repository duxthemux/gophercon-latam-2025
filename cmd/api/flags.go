package main

import (
	"log/slog"

	"github.com/urfave/cli/v3"
)

type flags struct {
	listeningAddr      string
	vecDbPath          string
	llmModel           string
	embModel           string
	minConfidenceRag   float64
	minConfidenceTool  float64
	minConfidenceCache float64
	temperature        float64
	toolDb             string

	tokenizerModel string
	tokenizerCache string

	slogLevel string
	otelEp    string
	llmEp     string
}

func (f *flags) SlogLevel() slog.Level {
	switch f.slogLevel {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelError
	}
}

func (f *flags) build() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "listening-addr",
			Value:       ":8080",
			Destination: &f.listeningAddr,
			DefaultText: ":8080",
			Sources:     cli.EnvVars("ADDR"),
		},
		&cli.StringFlag{
			Name:        "db",
			Value:       "./db",
			Destination: &f.vecDbPath,
			DefaultText: "./db",
			Sources:     cli.EnvVars("VEC_DB"),
		},
		&cli.StringFlag{
			Name:        "emb-model",
			Value:       "nomic-embed-text",
			Destination: &f.embModel,
			DefaultText: "nomic-embed-text",
			Sources:     cli.EnvVars("EMB_MODEL"),
		},
		&cli.StringFlag{
			Name:        "llm-model",
			Value:       "gemma3",
			Destination: &f.llmModel,
			DefaultText: "gemma3",
			Sources:     cli.EnvVars("LLM_MODEL"),
		},
		&cli.FloatFlag{
			Name:        "min-confidence-rag",
			Value:       0.80,
			Destination: &f.minConfidenceRag,
			DefaultText: "0.5",
			Sources:     cli.EnvVars("MIN_CONFIDENCE_RAG"),
		},
		&cli.FloatFlag{
			Name:        "min-confidence-tool",
			Value:       0.60,
			Destination: &f.minConfidenceTool,
			DefaultText: "0.6",
			Sources:     cli.EnvVars("MIN_CONFIDENCE_TOOL"),
		},
		&cli.FloatFlag{
			Name:        "min-confidence-cache",
			Value:       0.9,
			Destination: &f.minConfidenceCache,
			DefaultText: "0.9",
			Sources:     cli.EnvVars("MIN_CONFIDENCE_CACHE"),
		},
		&cli.FloatFlag{
			Name:        "temperature",
			Value:       0.5,
			Destination: &f.temperature,
			DefaultText: "0.5",
			Sources:     cli.EnvVars("TEMPERATURE"),
		},
		&cli.StringFlag{
			Name:        "slog-level",
			Value:       "info",
			Destination: &f.slogLevel,
			DefaultText: "info",
			Sources:     cli.EnvVars("SLOG_LEVEL"),
		},
		&cli.StringFlag{
			Name:        "otel-ep",
			Value:       "",
			Destination: &f.otelEp,
			DefaultText: "",
			Sources:     cli.EnvVars("OTEL_ENDPOINT"),
		},
		&cli.StringFlag{
			Name:        "tokenizer-model",
			Value:       "bert-base-uncased",
			Destination: &f.tokenizerModel,
			DefaultText: "bert-base-uncased",
			Sources:     cli.EnvVars("TOKENIZER_MODEL"),
		},
		&cli.StringFlag{
			Name:        "tokenizer-cache",
			Value:       "./data/tokenizer.json",
			Destination: &f.tokenizerCache,
			DefaultText: "./data/tokenizer.json",
			Sources:     cli.EnvVars("TOKENIZER_CACHE"),
		},
		&cli.StringFlag{
			Name:        "tool-db",
			Value:       "./data/db.sqlite",
			Destination: &f.toolDb,
			DefaultText: "./data/db.sqlite",
			Sources:     cli.EnvVars("TOOL_DB"),
		},
		&cli.StringFlag{
			Name:        "llm-ep",
			Value:       "localhost:11434",
			Destination: &f.llmEp,
			DefaultText: "localhost:11434",
			Sources:     cli.EnvVars("LLM_ENDPOINT"),
		},
	}
}
