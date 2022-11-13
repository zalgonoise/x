package log

import "context"

type Logger interface {
	Trace(msg string, attrs ...Attr)
	Debug(msg string, attrs ...Attr)
	Info(msg string, attrs ...Attr)
	Warn(msg string, attrs ...Attr)
	Error(msg string, attrs ...Attr)
	Fatal(msg string, attrs ...Attr)
	Log(level Level, msg string, attrs ...Attr)
	Enabled(level Level) bool
	Handler() Handler
	With(attrs ...Attr) Logger
}

func Default() Logger
func New(h Handler) Logger
func With(attrs ...Attr) Logger

func InContext(ctx context.Context, logger Logger) context.Context
func From(ctx context.Context) Logger
