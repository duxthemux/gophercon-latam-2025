package llm

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	ollama_api "github.com/ollama/ollama/api"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"gophercon-2025/cmd/api/cache"
	"gophercon-2025/cmd/api/rag"
	"gophercon-2025/cmd/api/tokenizer"
	"gophercon-2025/cmd/api/tool"
)

type Response struct {
	Type       string            `json:"type"`
	Response   string            `json:"response"`
	Tool       string            `json:"tool"`
	Params     map[string]string `json:"params"`
	Confidence float64           `json:"confidence"`
}

func cleanJson(s string) string {
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimPrefix(s, "json")
	s = strings.TrimSuffix(s, "json")
	s = strings.TrimSpace(s)

	return s
}

func (r *Response) LoadFromJson(s string) error {
	s = cleanJson(s)

	return json.Unmarshal([]byte(s), &r)
}

type Service struct {
	rag                *rag.Service
	cache              *cache.Service
	tokenizer          *tokenizer.Service
	tool               *tool.Service
	ollama             *ollama_api.Client
	logger             *slog.Logger
	tracer             trace.Tracer
	llmModel           string
	minConfidenceRag   float64
	minConfidenceTool  float64
	minConfidenceCache float64
	temperature        float64

	metricTokensInLlm    metric.Int64Counter
	metricTokensOutLlm   metric.Int64Counter
	metricTokensInCache  metric.Int64Counter
	metricTokensOutCache metric.Int64Counter
	metricCantAnswer     metric.Int64Counter
}

func (s *Service) Query(ctx context.Context, q string, useCache bool) (ret *Response, err error) {
	ctx, span := s.tracer.Start(ctx, "llm.Query")
	defer func() {
		span.RecordError(err)
		span.End()
	}()

	if useCache {
		response, err := s.checkCache(ctx, q)
		if err != nil {
			return nil, err
		}

		if response != "" {
			if err = s.addCacheMetrics(ctx, span, q, response); err != nil {
				return nil, err
			}

			return &Response{
				Type:     "FINAL",
				Response: response,
			}, nil
		}
	}

	ret, err = s.query(ctx, q)
	if err != nil {
		return nil, err
	}

	switch {
	case strings.HasSuffix(ret.Response, "\nRAG"):
		s.metricCantAnswer.Add(ctx, 1)
	case useCache && ret.Confidence > s.minConfidenceCache:
		if err = s.cache.Add(ctx, q, ret.Response, ""); err != nil {
			return nil, err
		}
	}

	return ret, nil
}

func New(options ...Option) *Service {
	ret := &Service{}

	for _, option := range options {
		option(ret)
	}

	return ret
}

//go:embed system.txt
var system string

func (s *Service) query(octx context.Context, q string) (*Response, error) {
	ctx, span := s.tracer.Start(octx, "llm.query")
	defer func() {
		span.End()
	}()

	s.logger.Debug("Querying RAG", "query", q)

	ragResSet, err := s.rag.Query(ctx, q)
	if err != nil {
		return nil, err
	}

	sb := strings.Builder{}

	prepareHeaderFunc := sync.OnceFunc(func() {
		sb.WriteString("Seu contexto contem dados do RAG. Considere as seguintes afirmações:\n")
	})

	span.AddEvent("rag returned", trace.WithAttributes(attribute.Int("results", len(ragResSet))))

	for _, ragRes := range ragResSet {
		if ragRes.Similarity > float32(s.minConfidenceTool) && ragRes.Metadata != nil && ragRes.Metadata["type"] == "TOOL" {
			ret, err := s.queryTool(ctx, q, ragRes.Metadata["name"])
			if err != nil {
				return nil, err
			}

			sb.WriteString(" - " + ret + "\n")

			continue
		}

		if ragRes.Similarity > float32(s.minConfidenceRag) {
			prepareHeaderFunc()
			s.logger.Debug("RAG: add responses", "query", q, "content", ragRes.Content, "similarity", ragRes.Similarity)
			sb.WriteString(" - " + ragRes.Content + "\n")
		}
	}

	if len(sb.String()) > 0 {
		sb.WriteString("Agora responda: " + q)
	} else {
		sb.WriteString(q)
	}

	ollamaReq := &ollama_api.GenerateRequest{
		Model:  s.llmModel,
		Prompt: sb.String(),
		System: system,
		Stream: new(bool),
		Options: map[string]any{
			"temperature": s.temperature,
		},
	}

	var ret *Response

	respFunc := func(resp ollama_api.GenerateResponse) error {
		svcResp := &Response{}
		if err := svcResp.LoadFromJson(resp.Response); err != nil {
			return err
		}

		ret = svcResp

		return nil
	}

	s.logger.Debug("Generating LLM response", "query", q)

	if err = s.ollama.Generate(ctx, ollamaReq, respFunc); err != nil {
		return nil, err
	}

	if err := s.addLlmMetrics(ctx, span, q, ret.Response); err != nil {
		return nil, err
	}

	return ret, nil
}

//go:embed llm_tool_prompt.txt
var llmToolPrompt string

func (s *Service) queryTool(octx context.Context, q string, tool string) (ret string, err error) {
	ctx, span := s.tracer.Start(octx, "llm.queryTool", trace.WithAttributes(attribute.String("q", q)))
	defer func() {
		span.RecordError(err)
		span.End()
	}()

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("Considere que a data de hoje é: %s\n", time.Now().String()))
	sb.WriteString(llmToolPrompt)
	sb.WriteString(q)

	ollamaReq := &ollama_api.GenerateRequest{
		Model:  s.llmModel,
		Prompt: sb.String(),
		Stream: new(bool),
		Options: map[string]any{
			"temperature": 0.2,
		},
	}

	respFunc := func(resp ollama_api.GenerateResponse) error {
		jsonStr := cleanJson(resp.Response)

		params := map[string]string{}
		if err = json.Unmarshal([]byte(jsonStr), &params); err != nil {
			return err
		}

		for k, v := range params {
			params[strings.ToLower(k)] = v
		}

		ret, err = s.tool.Query(ctx, tool, params)

		return err
	}

	if err = s.ollama.Generate(ctx, ollamaReq, respFunc); err != nil {
		return "", err
	}

	return ret, nil
}

func (s *Service) checkCache(ctx context.Context, q string) (ret string, err error) {
	ctx, span := s.tracer.Start(ctx, "llm.checkCache", trace.WithAttributes(attribute.String("q", q)))
	defer func() {
		span.RecordError(err)
		span.End()
	}()

	res, err := s.cache.Query(ctx, q)
	if err != nil {
		return "", err
	}

	var maxSim float32

	for i, ares := range res {
		if ares.Similarity > maxSim {
			maxSim = ares.Similarity
		}

		if ares.Similarity > float32(s.minConfidenceCache) {
			span.AddEvent("using cache", trace.WithAttributes(attribute.Int("i", i)))

			return ares.Metadata["RESPONSE"], nil
		}
	}

	span.SetAttributes(attribute.Float64("max-similarity", float64(maxSim)))

	span.AddEvent("no cache entries found")

	return "", nil
}

func (s *Service) addCacheMetrics(ctx context.Context, span trace.Span, in string, out string) error {
	tokensIn, err := s.tokenizer.Count(in)
	if err != nil {
		return err
	}

	s.metricTokensInCache.Add(ctx, int64(tokensIn))
	span.SetAttributes(attribute.Int64("tokens-in-cache", int64(tokensIn)))

	tokensOut, err := s.tokenizer.Count(out)
	if err != nil {
		return err
	}

	s.metricTokensOutCache.Add(ctx, int64(tokensOut))
	span.SetAttributes(attribute.Int64("tokens-out-cache", int64(tokensOut)))

	return nil
}

func (s *Service) addLlmMetrics(ctx context.Context, span trace.Span, in string, out string) error {
	tokensIn, err := s.tokenizer.Count(in)
	if err != nil {
		return err
	}

	s.metricTokensInLlm.Add(ctx, int64(tokensIn))
	span.SetAttributes(attribute.Int64("tokens-in-llm", int64(tokensIn)))

	tokensOut, err := s.tokenizer.Count(out)
	if err != nil {
		return err
	}

	s.metricTokensOutLlm.Add(ctx, int64(tokensOut))
	span.SetAttributes(attribute.Int64("tokens-out-llm", int64(tokensOut)))

	return nil
}
