package spanner

import "context"

type trace struct {
	spans  []Span
	parent SpanID
}

type Tracer interface {
	Start(ctx context.Context, name string) (context.Context, Span)
}
