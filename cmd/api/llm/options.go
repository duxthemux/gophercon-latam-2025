package llm

import (
	"log/slog"

	ollama_api "github.com/ollama/ollama/api"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"gophercon-2025/cmd/api/cache"
	"gophercon-2025/cmd/api/rag"
	"gophercon-2025/cmd/api/tokenizer"
	"gophercon-2025/cmd/api/tool"
)

type Option func(*Service)

func WithMinConfidenceRag(minConfidenceRag float64) Option {
	return func(s *Service) {
		s.minConfidenceRag = minConfidenceRag
	}
}

func WithMinConfidenceTool(minConfidenceTool float64) Option {
	return func(s *Service) {
		s.minConfidenceTool = minConfidenceTool
	}
}

func WithMinConfidenceCache(minConfidenceCache float64) Option {
	return func(s *Service) {
		s.minConfidenceCache = minConfidenceCache
	}
}

func WithLlmModel(llmModel string) Option {
	return func(s *Service) {
		s.llmModel = llmModel
	}
}

func WithOllama(c *ollama_api.Client) Option {
	return func(s *Service) {
		s.ollama = c
	}
}

func WithRag(r *rag.Service) Option {
	return func(s *Service) {
		s.rag = r
	}
}

func WithCache(c *cache.Service) Option {
	return func(s *Service) {
		s.cache = c
	}
}

func WithTemperature(t float64) Option {
	return func(s *Service) {
		s.temperature = t
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(s *Service) {
		s.logger = logger
	}
}

func WithTracer(tracer trace.Tracer) Option {
	return func(s *Service) {
		s.tracer = tracer
	}
}

func WithTokenizer(tk *tokenizer.Service) Option {
	return func(s *Service) {
		s.tokenizer = tk
	}
}

func WithMetrics(meter metric.Meter) Option {
	return func(s *Service) {
		var err error

		s.metricTokensInLlm, err = meter.Int64Counter("tokens_in_llm")
		if err != nil {
			panic(err)
		}

		s.metricTokensOutLlm, err = meter.Int64Counter("tokens_out_llm")
		if err != nil {
			panic(err)
		}

		s.metricTokensInCache, err = meter.Int64Counter("tokens_in_cache")
		if err != nil {
			panic(err)
		}

		s.metricTokensOutCache, err = meter.Int64Counter("tokens_out_cache")
		if err != nil {
			panic(err)
		}

		s.metricCantAnswer, err = meter.Int64Counter("cant_answer")
		if err != nil {
			panic(err)
		}
	}
}

func WithTool(tool *tool.Service) Option {
	return func(s *Service) {
		s.tool = tool
	}
}
