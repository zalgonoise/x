package spanner

import (
	"context"
)

type SpannerContextKey string

var ContextKey SpannerContextKey = "spanner"

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

func WithNewTrace(ctx context.Context) (context.Context, Trace) {
	t := newTrace()
	return context.WithValue(ctx, ContextKey, t), t
}

func WithTrace(ctx context.Context, t Trace) context.Context {
	return context.WithValue(ctx, ContextKey, t)
}
