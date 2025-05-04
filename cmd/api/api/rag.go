package api

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/philippgille/chromem-go"
)

type (
	ragClearRequest  struct{}
	ragClearResponse struct{}
)

func (a *Service) ragClear(ctx context.Context, req *ragClearRequest) (*ragClearResponse, error) {
	err := a.rag.Clear(ctx)

	return &ragClearResponse{}, err
}

type ragAddRequest struct {
	Body struct {
		Fact string            `json:"fact,omitempty"`
		Meta map[string]string `json:"meta,omitempty"`
	}
}
type ragAddResponse struct{}

func (a *Service) ragAdd(ctx context.Context, req *ragAddRequest) (*ragAddResponse, error) {
	err := a.rag.Add(ctx, req.Body.Fact, req.Body.Meta)

	return &ragAddResponse{}, err
}

type ragQueryRequest struct {
	Q string `json:"q" query:"q"`
	E bool   `json:"e" query:"e"`
}

type ragQueryResponse struct {
	Body []chromem.Result
}

func (a *Service) ragQuery(ctx context.Context, req *ragQueryRequest) (*ragQueryResponse, error) {
	docs, err := a.rag.Query(ctx, req.Q)
	if err != nil {
		return nil, err
	}

	if !req.E {
		for i := range docs {
			docs[i].Embedding = nil
		}
	}

	return &ragQueryResponse{Body: docs}, err
}

type ragDelRequest struct {
	Id string `path:"id"`
}

type ragDelResponse struct {
	Body []chromem.Result
}

func (a *Service) ragDel(ctx context.Context, req *ragDelRequest) (*ragDelResponse, error) {
	err := a.rag.Del(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &ragDelResponse{}, nil
}

func (a *Service) setupApiRag(humaApi huma.API) {
	huma.Register(humaApi, huma.Operation{
		OperationID: "apiV1RagOpClearPost",
		Method:      "POST",
		Path:        "/api/v1/rag/op/clear",
		Description: "Clears rag content",
	}, a.ragClear)

	huma.Register(humaApi, huma.Operation{
		OperationID: "apiV1RagGet",
		Method:      "GET",
		Path:        "/api/v1/rag",
		Description: "Queries rag",
	}, a.ragQuery)

	huma.Register(humaApi, huma.Operation{
		OperationID: "apiV1RagPost",
		Method:      "POST",
		Path:        "/api/v1/rag",
		Description: "Adds entry to rag",
	}, a.ragAdd)

	huma.Register(humaApi, huma.Operation{
		OperationID: "apiV1RagDelete",
		Method:      "DELETE",
		Path:        "/api/v1/rag/{id}",
		Description: "Dels entry from rag",
	}, a.ragDel)
}
