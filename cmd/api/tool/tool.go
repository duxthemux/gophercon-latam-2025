package tool

import (
	"context"
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
	"go.opentelemetry.io/otel/trace"
)

type Tool interface {
	Query(ctx context.Context, params map[string]string) (ret string, err error)
}

type Service struct {
	tools       map[string]Tool
	defaultTool Tool
	tracer      trace.Tracer
}

func (s *Service) Query(ctx context.Context, toolName string, params map[string]string) (ret string, err error) {
	ctx, span := s.tracer.Start(ctx, "tool.Query")
	defer func() {
		span.RecordError(err)
		span.End()
	}()

	params["tool"] = toolName

	tool, ok := s.tools[toolName]
	if !ok {
		return s.defaultTool.Query(ctx, params)
	}

	return tool.Query(ctx, params)
}

type Option func(service *Service)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func WithDb(db *sql.DB) Option {
	return func(service *Service) {
		goose.SetBaseFS(embedMigrations)

		if err := goose.SetDialect("postgres"); err != nil {
			panic(err)
		}

		if err := goose.Up(db, "migrations"); err != nil {
			panic(err)
		}

		service.defaultTool = &dbTool{db: db}
	}
}

func WithTracer(t trace.Tracer) Option {
	return func(service *Service) {
		service.tracer = t
	}
}

func New(opts ...Option) *Service {
	ret := &Service{
		tools: map[string]Tool{
			"df":       &diskFreeTool{},
			"hostname": &hostnameTool{},
			"ifconfig": &ifconfigTool{},
			"date":     &dateTool{},
		},
	}

	for _, opt := range opts {
		opt(ret)
	}

	return ret
}
