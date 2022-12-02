package spanner

import (
	"context"
)

// SpannerContextKey is a unique type to use as key when storing traces in context
type SpannerContextKey string

// ContextKey is the package's default context key for storing traces
var ContextKey SpannerContextKey = "spanner"

// GetTrace returns the Trace from the input context `ctx`, or nil if it doesn't have one
func GetTrace(ctx context.Context) Trace {
	v := ctx.Value(ContextKey)
	if v == nil {
		return nil
	}
	if t, ok := v.(Trace); ok {
		return t
	}
	return nil
}

// WithNewTrace wraps the input context `ctx` with a new Trace, returning both the
// context with value and the Trace
func WithNewTrace(ctx context.Context) (context.Context, Trace) {
	t := newTrace()
	return context.WithValue(ctx, ContextKey, t), t
}

// WithTrace wraps the input context `ctx` with the Trace `t`, returning the context with value
func WithTrace(ctx context.Context, t Trace) context.Context {
	return context.WithValue(ctx, ContextKey, t)
}
