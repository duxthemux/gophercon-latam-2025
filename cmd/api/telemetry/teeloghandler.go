package telemetry

import (
	"context"
	"log/slog"
)

type teeLogHandler struct {
	handlers []slog.Handler
	logLevel slog.Level
}

func (t teeLogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= t.logLevel
}

func (t teeLogHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, h := range t.handlers {
		if err := h.Handle(ctx, record); err != nil {
			return err
		}
	}

	return nil
}

func (t teeLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	ret := &teeLogHandler{handlers: make([]slog.Handler, len(t.handlers))}
	for i, h := range t.handlers {
		ret.handlers[i] = h.WithAttrs(attrs)
	}

	return ret
}

func (t teeLogHandler) WithGroup(name string) slog.Handler {
	ret := &teeLogHandler{handlers: make([]slog.Handler, len(t.handlers))}
	for i, h := range t.handlers {
		ret.handlers[i] = h.WithGroup(name)
	}

	return ret
}
