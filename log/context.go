package log

import "context"

type CtxLoggerKey string

const StandardCtxKey CtxLoggerKey = "logger"

func InContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, StandardCtxKey, logger)
}
func From(ctx context.Context) Logger {
	logger := ctx.Value(StandardCtxKey)
	if logger == nil {
		return std
	}
	return (logger).(Logger)
}
