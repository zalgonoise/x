package spanner

import (
	"context"

	"github.com/zalgonoise/logx/attr"
)

// Tracer will capture spans when its `Start()` method is called, by creating a new
// Trace or reusing the existing one in the input context `ctx`
//
// # Each call creates a Span which is appended to the Trace, and the Trace keeps running until the firstmost Span has ended
//
// The input context will not create a new SpanID, but reuse the previous. The returned context will use a new SpanID for the
// created Span, which is returned alongside this context.
//
// The returned Span is required, even if to defer its closure, with `defer s.End()`
type Tracer interface {
	Start(ctx context.Context, name string, attrs ...attr.Attr) (context.Context, Span)
}

type baseTracer struct{}

var tr Tracer = baseTracer{}

// Start reuses the Trace in the input context `ctx`, or creates one if it doesn't exist. It also
// creates the Span for the action, with string name `name` and Attr attributes `attrs`
//
// # Each call creates a Span which is appended to the Trace, and the Trace keeps running until the firstmost Span has ended
//
// The input context will not create a new SpanID, but reuse the previous. The returned context will use a new SpanID for the
// created Span, which is returned alongside this context.
//
// The returned Span is required, even if to defer its closure, with `defer s.End()`
func (baseTracer) Start(ctx context.Context, name string, attrs ...attr.Attr) (context.Context, Span) {
	ctx, t := GetTraceOrCreate(ctx)
	var pid *SpanID = nil
	parent := GetSpan(ctx)
	if parent != nil {
		id := parent.ID()
		pid = &id
	}

	s := newSpan(t.Receiver(), t.ID(), pid, name, attrs...)
	t.Register(s)

	ctx = WithTrace(ctx, t)
	ctx = WithSpan(ctx, t.Parent())
	newCtx := WithSpan(ctx, s)

	s.Start()
	return newCtx, s
}

// Start reuses the Trace in the input context `ctx`, or creates one if it doesn't exist. It also
// creates the Span for the action, with string name `name` and Attr attributes `attrs`
//
// # The created Span is appended to the Trace, and the Trace keeps running until the firstmost Span has ended
//
// The input context will not create a new SpanID, but reuse the previous. The returned context will use a new SpanID for the
// created Span, which is returned alongside this context.
//
// The returned Span is required, even if to defer its closure, with `defer s.End()`
func Start(ctx context.Context, name string, attrs ...attr.Attr) (context.Context, Span) {
	return tr.Start(ctx, name, attrs...)
}
