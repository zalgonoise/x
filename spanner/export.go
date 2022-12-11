package spanner

import (
	"fmt"
	"io"
)

// Exporter is a module that pushes the span data to an output
type Exporter interface {
	// Export pushes the input SpanData `spans` to its output and returns
	// an error
	Export(traceID TraceID, spans ...*SpanData) error
}

type noOpExporter struct{}

func (noOpExporter) Export(traceID TraceID, spans ...*SpanData) error {
	return nil
}

type writerExporter struct {
	enc Encoder
	w   io.Writer
}

func (e writerExporter) Export(traceID TraceID, spans ...*SpanData) error {
	trace := struct {
		TraceID string      `json:"trace_id"`
		Spans   []*SpanData `json:"spans"`
	}{
		TraceID: traceID.String(),
		Spans:   spans,
	}

	b, err := e.enc.Encode(trace)
	if err != nil {
		return err
	}
	fmt.Fprint(e.w, string(b))

	return nil
}

func Writer(w io.Writer) Exporter {
	return writerExporter{
		enc: jsonEnc{},
		w:   w,
	}
}
