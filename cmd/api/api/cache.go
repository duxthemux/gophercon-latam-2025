package api

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/philippgille/chromem-go"
)

type (
	cacheClearRequest  struct{}
	cacheClearResponse struct{}
)

func (a *Service) cacheClear(ctx context.Context, req *cacheClearRequest) (*cacheClearResponse, error) {
	err := a.cache.Clear(ctx)

	return &cacheClearResponse{}, err
}

type cacheAddRequest struct {
	Body cacheAddRequestBody
}
type cacheAddRequestBody struct {
	Fact     string `json:"fact"`
	Meta     string `json:"meta"`
	Response string `json:"response"`
}
type cacheAddResponse struct{}

func (a *Service) cacheAdd(ctx context.Context, req *cacheAddRequest) (*cacheAddResponse, error) {
	err := a.cache.Add(ctx, req.Body.Fact, req.Body.Response, req.Body.Meta)

	return &cacheAddResponse{}, err
}

type cacheQueryRequest struct {
	Q string `json:"q" query:"q"`
	E bool   `json:"e" query:"e"`
}

type cacheQueryResponse struct {
	Body []chromem.Result
}

func (a *Service) cacheQuery(ctx context.Context, req *cacheQueryRequest) (*cacheQueryResponse, error) {
	docs, err := a.cache.Query(ctx, req.Q)
	if err != nil {
		return nil, err
	}

	if !req.E {
		for i := range docs {
			docs[i].Embedding = nil
		}
	}

	return &cacheQueryResponse{Body: docs}, err
}

type cacheDelRequest struct {
	Id string `path:"id"`
}

type cacheDelResponse struct {
	Body []chromem.Result
}

func (a *Service) cacheDel(ctx context.Context, req *cacheDelRequest) (*cacheDelResponse, error) {
	err := a.cache.Del(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &cacheDelResponse{}, nil
}

func (a *Service) setupApiCache(humaApi huma.API) {
	huma.Register(humaApi, huma.Operation{
		OperationID: "apiV1CacheOpClearPost",
		Method:      "POST",
		Path:        "/api/v1/cache/op/clear",
		Description: "Clears cache content",
	}, a.cacheClear)

	huma.Register(humaApi, huma.Operation{
		OperationID: "apiV1CacheGet",
		Method:      "GET",
		Path:        "/api/v1/cache",
		Description: "Queries cache",
	}, a.cacheQuery)

	huma.Register(humaApi, huma.Operation{
		OperationID: "apiV1CachePost",
		Method:      "POST",
		Path:        "/api/v1/cache",
		Description: "Adds entry to cache",
	}, a.cacheAdd)

	huma.Register(humaApi, huma.Operation{
		OperationID: "apiV1CacheDelete",
		Method:      "DELETE",
		Path:        "/api/v1/cache/{id}",
		Description: "Dels entry from cache",
	}, a.cacheDel)
}
