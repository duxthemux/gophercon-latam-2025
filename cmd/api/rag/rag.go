package rag

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/philippgille/chromem-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	ColletionNameRag      = "rag"
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

func WithDBName(dbPath string) Option {
	if dbPath == "" {
		panic("dbPath is empty")
	}

	return func(s *Service) {
		db, err := chromem.NewPersistentDB(dbPath, true)
		if err != nil {
			panic(err)
		}

		s.db = db
	}
}

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

func WithTracer(tracer trace.Tracer) Option {
	return func(s *Service) {
		s.tracer = tracer
	}
}

func WithOllamaEndpoint(endpoint string) Option {
	return func(s *Service) {
		s.llmEp = endpoint
	}
}

func (r *Service) Add(ctx context.Context, fact string, meta map[string]string) (err error) {
	ctx, span := r.tracer.Start(ctx, "rag.Add")
	defer func() {
		span.RecordError(err)
		span.End()
	}()

	err = r.collection().AddDocument(ctx, chromem.Document{
		ID:       uuid.NewString(),
		Metadata: meta,
		Content:  fact,
	})

	return err
}

func (r *Service) Clear(ctx context.Context) (err error) {
	_, span := r.tracer.Start(ctx, "rag.Clear")
	defer func() {
		span.RecordError(err)
		span.End()
	}()

	err = r.db.DeleteCollection(ColletionNameRag)

	return err
}

func (r *Service) Query(ctx context.Context, s string) (ret []chromem.Result, err error) {
	ctx, span := r.tracer.Start(ctx, "rag.Query", trace.WithAttributes(attribute.String("query", s)))
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
	ctx, span := r.tracer.Start(ctx, "rag.Del")
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

	defaultEmbeddingFunc = chromem.NewEmbeddingFuncOllama(ret.embModel, ret.llmEp+"/api")

	if ret.db == nil {
		return nil, errors.New("db was not initialized")
	}

	return ret, nil
}
