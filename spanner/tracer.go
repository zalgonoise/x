package spanner

import (
	"context"
	"os"

	"github.com/zalgonoise/attr"
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

type baseTracer struct {
	e Exporter
}

var tr Tracer = &baseTracer{
	e: Writer(os.Stderr),
}

// Start reuses the Trace in the input context `ctx`, or creates one if it doesn't exist. It also
// creates the Span for the action, with string name `name` and Attr attributes `attrs`
//
// # Each call creates a Span which is appended to the Trace, and the Trace keeps running until the firstmost Span has ended
//
// The input context will not create a new SpanID, but reuse the previous. The returned context will use a new SpanID for the
// created Span, which is returned alongside this context.
//
// The returned Span is required, even if to defer its closure, with `defer s.End()`
func (t baseTracer) Start(ctx context.Context, name string, attrs ...attr.Attr) (context.Context, Span) {
	var trace Trace
	ctx, trace = GetTraceOrCreate(ctx, t.e)

	parent := GetSpan(ctx)
	trace.Register(parent)

	s := newSpan(trace, name, attrs...)
	trace.Add(s)

	ctx = WithTrace(ctx, trace)
	ctx = WithSpan(ctx, trace.Parent())
	trace.Register(s)
	newCtx := WithSpan(ctx, s)

	s.Start()
	return newCtx, s
}

func (t *baseTracer) To(e Exporter) {
	if e == nil {
		t.e = noOpExporter{}
	}
	t.e = e
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

// To globally sets the Span exporter to Exporter `e`
func To(e Exporter) {
	tr.(*baseTracer).To(e)
}
