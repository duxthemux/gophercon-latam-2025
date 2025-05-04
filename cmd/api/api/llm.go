package api

import (
	"context"
	_ "embed"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

type llmQueryRequest struct {
	Body struct {
		Query    string `json:"query,omitempty"`
		Details  bool   `json:"details,omitempty"`
		UseCache bool   `json:"use_cache,omitempty"`
	}
}
type llmQueryResponse struct {
	Body any
}

func (a *Service) llmQuery(ctx context.Context, req *llmQueryRequest) (*llmQueryResponse, error) {
	start := time.Now()
	defer func() {
		dur := time.Since(start)
		a.metricResponseTime.Add(ctx, dur.Seconds())
	}()

	ret, err := a.llm.Query(ctx, req.Body.Query, req.Body.UseCache)
	if err != nil {
		return nil, err
	}

	if req.Body.Details {
		return &llmQueryResponse{Body: ret}, nil
	}

	return &llmQueryResponse{Body: ret.Response}, nil
}

func (a *Service) setupApiLlm(humaApi huma.API) {
	huma.Register(humaApi, huma.Operation{
		OperationID: "apiV1LlmQueryPost",
		Method:      "POST",
		Path:        "/api/v1/llm",
		Description: "retrieves general status of this service",
	}, a.llmQuery)
}
