package spanner

// Trace records and stores the set of Spans, as actions in a transaction
//
// It exposes methods for retrieving the slice of Spans, adding a new Span,
// retrieving the TraceID, and retrieving the next span's parent SpanID if set
type Trace interface {
	// Get returns the slice of Spans in the Trace
	Get() []Span
	// Add takes in two Spans, the first one which is appended to the list of Spans
	// and the second one which is used as a reference for the parent SpanID in the
	// next Trace
	//
	// Traces are immutable, and adding a Span returns a new Trace
	Add(s, ref Span) Trace
	// ID returns the TraceID
	ID() TraceID
	// PID returns the parent SpanID, or nil if unset
	PID() *SpanID
}

type trace struct {
	spans []Span
	trace TraceID
	ref   Span
}

func newTrace() Trace {
	return trace{
		spans: []Span{},
		trace: NewTraceID(),
	}
}

// Get returns the slice of Spans in the Trace
func (t trace) Get() []Span {
	return t.spans
}

// Add takes in two Spans, the first one which is appended to the list of Spans
// and the second one which is used as a reference for the parent SpanID in the
// next Trace
//
// Traces are immutable, and adding a Span returns a new Trace
func (t trace) Add(s, ref Span) Trace {
	sCopy := make([]Span, 0, len(t.spans))
	copy(sCopy, t.spans)

	sCopy = append(sCopy, s)

	return trace{
		spans: sCopy,
		trace: t.trace,
		ref:   ref,
	}
}

// ID returns the TraceID
func (t trace) ID() TraceID {
	return t.trace
}

// PID returns the parent SpanID, or nil if unset
func (t trace) PID() *SpanID {
	if t.ref == nil {
		return nil
	}
	id := t.ref.ID()
	return &id
}
