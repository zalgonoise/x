package spanner

import (
	"context"
	"fmt"
	"io"
)

// Exporter is a module that pushes the span data to an output
type Exporter interface {
	// Export pushes the input SpanData `spans` to its output, as a non-blocking
	// function
	Export(ctx context.Context, spans []SpanData) error
	Shutdown(ctx context.Context) error
}

type noOpExporter struct{}

func (noOpExporter) Export(ctx context.Context, spans []SpanData) error {
	return nil
}
func (noOpExporter) Shutdown(ctx context.Context) error {
	return nil
}

type writerExporter struct {
	enc Encoder
	w   io.Writer
}

func (e writerExporter) Export(ctx context.Context, spans []SpanData) error {
	var exportErr error
	for _, span := range spans {
		b, _ := span.MarshalJSON()
		if _, err := e.w.Write(b); err != nil {
			if exportErr == nil {
				exportErr = err
			} else {
				exportErr = fmt.Errorf("%w -- %v", err, exportErr)
			}
		}
	}

	return exportErr
}

func (e writerExporter) Shutdown(ctx context.Context) error {
	if wc, ok := e.w.(interface {
		Close() error
	}); ok {
		return wc.Close()
	}
	return nil
}

func Writer(w io.Writer) Exporter {
	return writerExporter{
		enc: jsonEnc{},
		w:   w,
	}
}
