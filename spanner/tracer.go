package spanner

import (
	"context"

	"github.com/zalgonoise/logx/attr"
)

type Tracer interface {
	Start(ctx context.Context, name string, attrs ...attr.Attr) (context.Context, Span)
}

type baseTracer struct{}

var tr Tracer = baseTracer{}

func (baseTracer) Start(ctx context.Context, name string, attrs ...attr.Attr) (context.Context, Span) {
	t := GetTrace(ctx)
	if t == nil {
		ctx, _ = WithNewTrace(ctx)
	}

	newCtx, s := addSpan(ctx, name, attrs...)
	s.Start()
	return newCtx, s
}

func Start(ctx context.Context, name string, attrs ...attr.Attr) (context.Context, Span) {
	return tr.Start(ctx, name, attrs...)
}

func addSpan(ctx context.Context, name string, attrs ...attr.Attr) (context.Context, Span) {
	t := GetTrace(ctx)
	if t == nil {
		ctx, t = WithNewTrace(ctx)
	}
	s := newSpan(t, name, attrs...)

	set := t.Add(s, s)
	unset := t.Add(s, nil)

	ctx = WithTrace(ctx, unset)
	newCtx := WithTrace(ctx, set)

	return newCtx, s
}
