package spanner

type Trace interface {
	Get() []Span
	Add(s Span, sid Span) Trace
	ID() TraceID
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

func (t trace) Get() []Span {
	return t.spans
}
func (t trace) Add(s Span, ref Span) Trace {
	sCopy := make([]Span, 0, len(t.spans))
	copy(sCopy, t.spans)

	sCopy = append(sCopy, s)

	return trace{
		spans: sCopy,
		trace: t.trace,
		ref:   ref,
	}
}

func (t trace) ID() TraceID {
	return t.trace
}

func (t trace) PID() *SpanID {
	if t.ref == nil {
		return nil
	}
	id := t.ref.ID()
	return &id
}
