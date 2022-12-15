package spanner

import (
	"context"
)

// TraceContextKey is a unique type to use as key when storing traces in context
type TraceContextKey string

// ContextKey is the package's default context key for storing traces in context
var ContextKey TraceContextKey = "spanner"

// SpanIDContextKey is a unique type to use as key when storing span IDs in context,
// which are referenced as parent Span IDs in new Spans
type SpanIDContextKey string

// SpanContextKey is the package's default context key for storing Spans in context
var SpanContextKey SpanIDContextKey = "span"

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

// GetSpan returns the Span from the input context `ctx`, or nil if it doesn't have one
func GetSpan(ctx context.Context) Span {
	v := ctx.Value(SpanContextKey)
	if v == nil {
		return nil
	}
	if s, ok := v.(Span); ok {
		return s
	}
	return nil
}

// WithSpan wraps the input context `ctx` with the Span `s`, returning the context with value
func WithSpan(ctx context.Context, s Span) context.Context {
	return context.WithValue(ctx, SpanContextKey, s)
}
