package prom

import (
	"context"
	"log/slog"
)

type noOpLogHandler struct{}

func (noOpLogHandler) Enabled(context.Context, slog.Level) bool {
	return false
}

func (noOpLogHandler) Handle(context.Context, slog.Record) error {
	return nil
}

func (h noOpLogHandler) WithAttrs([]slog.Attr) slog.Handler {
	return h
}

func (h noOpLogHandler) WithGroup(string) slog.Handler {
	return h
}
