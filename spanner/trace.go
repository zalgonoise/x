package spanner

// Trace records and stores the set of Spans, as actions in a transaction
//
// It exposes methods for retrieving the slice of Spans, adding a new Span,
// retrieving the TraceID, and retrieving the next span's parent SpanID if set
type Trace interface {
	// Add appends a Span to the Trace
	Add(s Span)
	// Get returns the slice of Spans in the Trace
	Get() []Span
	// Register sets the input Span `s`'s SpanID as this Trace's reference parent_id
	Register(s Span)
	// ID returns the TraceID
	ID() TraceID
	// Parent returns the parent SpanID, or nil if unset
	Parent() Span
	// Extract returns the SpanData from the Trace, if any
	Extract() []SpanData
	// Export pushes the current slice of SpanData to the configured Exporter
	Export()
}

type trace struct {
	exporter Exporter
	trace    TraceID
	ref      Span
	spans    []Span
}

func newTrace(e Exporter) Trace {
	if e == nil {
		e = noOpExporter{}
	}
	t := &trace{
		exporter: e,
		trace:    NewTraceID(),
		spans:    []Span{},
	}
	return t
}

// Add appends a Span to the Trace
func (t *trace) Add(s Span) {
	t.spans = append(t.spans, s)
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

// Extract returns the SpanData from the Trace, if any
func (t trace) Extract() []SpanData {
	if len(t.spans) == 0 {
		return nil
	}
	data := make([]SpanData, 0, len(t.spans))
	for _, s := range t.spans {
		data = append(data, s.Extract())
	}
	return data
}

func (t *trace) Export() {
	t.exporter.Export(t)
}
