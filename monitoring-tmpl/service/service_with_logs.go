package service

import (
	"context"
	"log/slog"
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
	h.logger.InfoContext(ctx, "received a new value to process", slog.Int("value", value))

	if err = h.s.Handle(ctx, value); err != nil {
		h.logger.ErrorContext(ctx, "failed to handle the input value",
			slog.Int("value", value),
			slog.String("error", err.Error()),
		)

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
