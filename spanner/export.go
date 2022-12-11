package spanner

import (
	"io"
)

// Exporter is a module that pushes the span data to an output
type Exporter interface {
	// Export pushes the input SpanData `spans` to its output, as a non-blocking
	// function
	Export(trace Trace)
}

type noOpExporter struct{}

func (noOpExporter) Export(trace Trace) {}

type writerExporter struct {
	enc Encoder
	w   io.Writer
}

func (e writerExporter) Export(trace Trace) {
	tr := struct {
		TraceID string     `json:"trace_id"`
		Spans   []SpanData `json:"spans"`
	}{
		TraceID: trace.ID().String(),
		Spans:   trace.Extract(),
	}

	b, _ := e.enc.Encode(tr)
	_, _ = e.w.Write(b)
}

func Writer(w io.Writer) Exporter {
	return writerExporter{
		enc: jsonEnc{},
		w:   w,
	}
}

type rawExporter struct {
	w io.Writer
}

func (e rawExporter) Export(trace Trace) {
	sp := trace.Extract()

	for _, s := range sp {
		b, _ := s.MarshalJSON()
		_, _ = e.w.Write(b)
	}
}

func Raw(w io.Writer) Exporter {
	return rawExporter{
		w: w,
	}
}
