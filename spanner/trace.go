package spanner

import "sync/atomic"

// Trace records and stores the set of Spans, as actions in a transaction
//
// It exposes methods for retrieving the slice of Spans, adding a new Span,
// retrieving the TraceID, and retrieving the next span's parent SpanID if set
type Trace interface {
	// Register sets the input Span `s`'s SpanID as this Trace's reference parent_id
	Register(s *SpanID)
	// ID returns the TraceID
	ID() TraceID
	// Parent returns the parent SpanID, or nil if unset
	Parent() *SpanID
	// Tracer returns the configured Tracer in the Trace
	Processor() SpanProcessor
}

type trace struct {
	processor atomic.Value
	trace     TraceID
	ref       *SpanID
}

func newTrace(processor atomic.Value) Trace {
	newTr := &trace{
		processor: processor,
		trace:     NewTraceID(),
	}
	return newTr
}

// Add takes in two Spans, the first one which is appended to the list of Spans
// and the second one which is used as a reference for the parent SpanID in the
// next Trace
//
// Traces are immutable, and adding a Span returns a new Trace
func (t *trace) Register(s *SpanID) {
	t.ref = s
}

// ID returns the TraceID
func (t trace) ID() TraceID {
	return t.trace
}

// PID returns the parent SpanID, or nil if unset
func (t trace) Parent() *SpanID {
	return t.ref
}

func (t *trace) Processor() SpanProcessor {
	return t.processor.Load().(SpanProcessor)
}
