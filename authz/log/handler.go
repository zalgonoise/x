package log

import (
	"context"
	"log/slog"
)

type NoOp struct{}

func (NoOp) Enabled(context.Context, slog.Level) bool  { return false }
func (NoOp) Handle(context.Context, slog.Record) error { return nil }
func (h NoOp) WithAttrs([]slog.Attr) slog.Handler      { return h }
func (h NoOp) WithGroup(string) slog.Handler           { return h }
