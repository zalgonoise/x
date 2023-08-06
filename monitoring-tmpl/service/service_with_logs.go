package service

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type Logger interface {
	DebugContext(ctx context.Context, msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
}

var _ Service = HandlerWithLogs{}

type HandlerWithLogs struct {
	s      Service
	logger Logger
}

func (h HandlerWithLogs) Handle(ctx context.Context, value int) (err error) {
	attrs := make([]any, 0, 6)
	attrs = append(attrs, "value", value)

	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		attrs = append(attrs, "traceID", sc.TraceID().String())
	}

	h.logger.InfoContext(ctx, "received a new value to process", attrs...)

	if err = h.s.Handle(ctx, value); err != nil {
		attrs = append(attrs, "error", err.Error())
		h.logger.ErrorContext(ctx, "failed to handle the input value", attrs...)

		return err
	}

	return nil
}

func WithLogs(s Service, logger Logger) HandlerWithLogs {
	return HandlerWithLogs{
		s:      s,
		logger: logger,
	}
}
