package cache

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/philippgille/chromem-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	ColletionNameRag      = "cache"
	EmbeddingModel        = "nomic-embed-text"
	DefaultOllamaEndpoint = ""
)

var defaultEmbeddingFunc = chromem.NewEmbeddingFuncOllama(EmbeddingModel, DefaultOllamaEndpoint)

type Service struct {
	db     *chromem.DB
	tracer trace.Tracer

	embModel string
	llmEp    string
}

type Option func(*Service)

func WithDb(db *chromem.DB) Option {
	return func(s *Service) {
		s.db = db
	}
}

func WithEmbModel(model string) Option {
	return func(s *Service) {
		s.embModel = model
	}
}

func WithOllamaEndpoint(endpoint string) Option {
	return func(s *Service) {
		s.llmEp = endpoint
	}
}

func WithTracer(tracer trace.Tracer) Option {
	return func(s *Service) {
		s.tracer = tracer
	}
}

func (r *Service) Add(ctx context.Context, fact string, response, meta string) (err error) {
	ctx, span := r.tracer.Start(ctx, "cache.Add")
	defer func() {
		span.RecordError(err)
		span.End()
	}()

	metaMap := map[string]string{
		"RESPONSE": response,
	}

	entries := strings.Split(meta, ",")

	for _, entry := range entries {
		kv := strings.Split(entry, ":")
		if len(kv) != 2 || kv[0] == "RESPONSE" {
			continue
		}

		metaMap[kv[0]] = kv[1]
	}

	err = r.collection().AddDocument(ctx, chromem.Document{
		ID:       uuid.NewString(),
		Metadata: metaMap,
		Content:  fact,
	})

	return err
}

func (r *Service) Clear(ctx context.Context) (err error) {
	_, span := r.tracer.Start(ctx, "cache.Clear")
	defer func() {
		span.RecordError(err)
		span.End()
	}()

	err = r.db.DeleteCollection(ColletionNameRag)

	return err
}

func (r *Service) Query(ctx context.Context, s string) (ret []chromem.Result, err error) {
	ctx, span := r.tracer.Start(ctx, "cache.Query", trace.WithAttributes(attribute.String("query", s)))
	defer func() {
		span.RecordError(err)
		span.End()
	}()

	if r.collection().Count() < 1 {
		return nil, nil
	}

	nresults := 25
	if nresults > r.collection().Count() {
		nresults = r.collection().Count()
	}

	ret, err = r.collection().Query(ctx, s, nresults, nil, nil)

	return ret, err
}

func (r *Service) Del(ctx context.Context, id string) (err error) {
	ctx, span := r.tracer.Start(ctx, "cache.Del")
	defer func() {
		span.RecordError(err)
		span.End()
	}()

	err = r.collection().Delete(ctx, nil, nil, id)

	return err
}

func (r *Service) collection() *chromem.Collection {
	col := r.db.GetCollection(ColletionNameRag, defaultEmbeddingFunc)
	if col == nil {
		var err error

		col, err = r.db.CreateCollection(ColletionNameRag, nil, defaultEmbeddingFunc)
		if err != nil {
			panic(err)
		}
	}

	return col
}

func New(opts ...Option) (ret *Service, err error) {
	ret = &Service{}

	for _, opt := range opts {
		opt(ret)
	}

	if ret.db == nil {
		return nil, errors.New("db was not initialized")
	}

	defaultEmbeddingFunc = chromem.NewEmbeddingFuncOllama(ret.embModel, ret.llmEp+"/api")

	return ret, nil
}
