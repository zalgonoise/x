package spanner

import (
	"context"
	"os"
	"sync/atomic"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/logx"
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
	Start(ctx context.Context, name string) (context.Context, Span)
}

type baseTracer struct{}

var (
	exp  Exporter = Writer(os.Stderr)
	proc          = atomic.Value{}
	tr   Tracer   = &baseTracer{}
)

func init() {
	proc.Store(NewProcessor(exp))
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
func (t baseTracer) Start(ctx context.Context, name string) (context.Context, Span) {
	var trace Trace = GetTrace(ctx)
	if trace == nil {
		ctx, trace = WithNewTrace(ctx)
	}
	parent := GetSpan(ctx)

	s := newSpan(trace, name)
	sid := s.ID()

	ctx = WithTrace(ctx, trace)
	ctx = WithSpan(ctx, parent)
	trace.Register(&sid)
	newCtx := WithSpan(ctx, s)

	s.Start()
	return newCtx, s
}

func (t *baseTracer) To(e Exporter) {
	if e == nil {
		e = noOpExporter{}
	}

	p := proc.Load().(SpanProcessor)
	err := p.Shutdown(context.Background())
	if err != nil {
		logx.Error("failed to stop processor", attr.String("error", err.Error()))
	}

	proc.Store(NewProcessor(e))
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
func Start(ctx context.Context, name string) (context.Context, Span) {
	return tr.Start(ctx, name)
}

// To globally sets the Span exporter to Exporter `e`
func To(e Exporter) {
	tr.(*baseTracer).To(e)
}
