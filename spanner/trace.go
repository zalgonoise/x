package spanner

// Trace represents a single transaction which creates a set of Spans, as single-event actions.
//
// It exposes methods for registering and retrieving the parent SpanID to use in the next
// Tracer's `Start()` call, and for returning its TraceID.
type Trace interface {
	// ID returns the TraceID
	ID() TraceID
	// Register sets the input pointer to a SpanID `s` as this Trace's reference parent_id
	Register(s *SpanID)
	// Parent returns the parent SpanID, or nil if unset
	Parent() *SpanID
}

type trace struct {
	trace TraceID
	ref   *SpanID
}

func newTrace() Trace {
	newTr := &trace{
		trace: NewTraceID(),
	}
	return newTr
}

// Register sets the input pointer to a SpanID `s` as this Trace's reference parent_id
func (t *trace) Register(s *SpanID) {
	t.ref = s
}

// ID returns the TraceID
func (t trace) ID() TraceID {
	return t.trace
}

// Parent returns the parent SpanID, or nil if unset
func (t trace) Parent() *SpanID {
	return t.ref
}
