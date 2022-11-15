package log

import "context"

type CtxLoggerKey string

const StandardCtxKey CtxLoggerKey = "logger"

func InContext(ctx context.Context, logger Logger) context.Context {
	if ctx == nil || logger == nil {
		return nil
	}
	return context.WithValue(ctx, StandardCtxKey, logger)
}
func From(ctx context.Context) Logger {
	v := ctx.Value(StandardCtxKey)
	if l, ok := v.(Logger); ok {
		return l
	}
	return nil
}
