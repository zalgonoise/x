package spanner

import "fmt"

// Exporter is a module that pushes the span data to an output
type Exporter interface {
	// Export pushes the input SpanData `spans` to its output and returns
	// an error
	Export(traceID TraceID, spans ...SpanData) error
}

type ttyExporter struct {
	enc Encoder
}

func (e ttyExporter) Export(traceID TraceID, spans ...SpanData) error {
	trace := struct {
		TraceID string     `json:"trace_id"`
		Spans   []SpanData `json:"spans"`
	}{
		TraceID: traceID.String(),
		Spans:   spans,
	}

	b, err := e.enc.Encode(trace)
	if err != nil {
		return err
	}
	fmt.Println(string(b))

	return nil
}

func TTY() Exporter {
	return ttyExporter{
		enc: jsonEnc{},
	}
}
