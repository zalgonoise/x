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
	Register(s Span)
	// ID returns the TraceID
	ID() TraceID
	// PID returns the parent SpanID, or nil if unset
	Parent() Span
	// Receiver returns the Span receiving channel of the Tracer
	Receiver() chan Span
	// Extract returns the SpanData from the Trace, if any
	Extract() []SpanData
}

type trace struct {
	spans []Span
	trace TraceID
	ref   Span
	rcv   chan Span
}

func newTrace() Trace {
	t := &trace{
		spans: []Span{},
		trace: NewTraceID(),
		rcv:   make(chan Span),
	}
	go func() {
		for s := range t.rcv {
			t.spans = append(t.spans, s)
		}
	}()
	return t
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
func (t *trace) Register(s Span) {
	t.ref = s
}

// ID returns the TraceID
func (t trace) ID() TraceID {
	return t.trace
}

// PID returns the parent SpanID, or nil if unset
func (t trace) Parent() Span {
	return t.ref
}

// Receiver returns the Span receiving channel of the Tracer
func (t trace) Receiver() chan Span {
	return t.rcv
}

// Extract returns the SpanData from the Trace, if any
func (t trace) Extract() []SpanData {
	if len(t.spans) == 0 {
		return nil
	}
	data := make([]SpanData, len(t.spans))
	for idx, s := range t.spans {
		data[idx] = s.Extract()
	}
	return data
}
