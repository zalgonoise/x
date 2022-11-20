package log

import "context"

// CtxLoggerKey is a custom type to define context keys for this
// library's logger
type CtxLoggerKey string

// StandardCtxKey is an instance of CtxLoggerKey with value "logger"
const StandardCtxKey CtxLoggerKey = "logger"

// InContext returns a copy of the input Context `ctx` with the input
// Logger `logger` as a value (identified by `StandardCtxKey`)
func InContext(ctx context.Context, logger Logger) context.Context {
	if ctx == nil || logger == nil {
		return nil
	}
	return context.WithValue(ctx, StandardCtxKey, logger)
}

// From returns a Logger from the input Context `ctx`. If not present,
// it returns nil
func From(ctx context.Context) Logger {
	v := ctx.Value(StandardCtxKey)
	if l, ok := v.(Logger); ok {
		return l
	}
	return nil
}
