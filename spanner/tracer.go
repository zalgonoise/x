package spanner

import (
	"context"
	"os"
	"sync/atomic"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/logx"
)

// Tracer is responsible of creating new Traces and Spans, but it also allows to define the Exporter
// it should use
type Tracer interface {
	// Start reuses the Trace in the input context `ctx`, or creates one if it doesn't exist. It also
	// creates the Span for the action, with string name `name`. Each call creates a new Span.
	//
	// After calling Start, the input context will still reference the parent Span's ID, nil if it's a new Trace.
	// The returned context will reference the returned Span's ID, to be used as the next call's parent.
	//
	// The returned Span is required, even if to defer its closure, with `defer s.End()`. The caller MUST close the
	// returned Span.
	Start(ctx context.Context, name string) (context.Context, Span)
	// To sets the Span exporter to Exporter `e`
	To(e Exporter)
	// Processor returns the configured SpanProcessor in the Tracer
	Processor() SpanProcessor
}

type baseTracer struct{}

var (
	exp  Exporter = Writer(os.Stderr)
	proc          = atomic.Value{}
	tr   Tracer   = &baseTracer{}
)

func init() {
	// initialize a Span Exporter to os.Stderr
	proc.Store(NewProcessor(exp))
}

// Start reuses the Trace in the input context `ctx`, or creates one if it doesn't exist. It also
// creates the Span for the action, with string name `name`. Each call creates a new Span.
//
// After calling Start, the input context will still reference the parent Span's ID, nil if it's a new Trace.
// The returned context will reference the returned Span's ID, to be used as the next call's parent.
//
// The returned Span is required, even if to defer its closure, with `defer s.End()`. The caller MUST close the
// returned Span.
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

// To sets the Span exporter to Exporter `e`
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

func (t *baseTracer) Processor() SpanProcessor {
	return proc.Load().(SpanProcessor)
}

// Start reuses the Trace in the input context `ctx`, or creates one if it doesn't exist. It also
// creates the Span for the action, with string name `name`. Each call creates a new Span.
//
// After calling Start, the input context will still reference the parent Span's ID, nil if it's a new Trace.
// The returned context will reference the returned Span's ID, to be used as the next call's parent.
//
// The returned Span is required, even if to defer its closure, with `defer s.End()`. The caller MUST close the
// returned Span.
func Start(ctx context.Context, name string) (context.Context, Span) {
	return tr.Start(ctx, name)
}

// To globally sets the Span exporter to Exporter `e`
func To(e Exporter) {
	tr.(*baseTracer).To(e)
}

func Processor() SpanProcessor {
	return tr.Processor()
}
