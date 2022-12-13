package spanner

import (
	"context"
	"fmt"
	"io"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/logx"
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
	rcv  chan []SpanData
	done chan struct{}
	enc  Encoder
	w    io.Writer
}

func (e writerExporter) Export(ctx context.Context, spans []SpanData) error {
	e.rcv <- spans
	return nil
}

func (e writerExporter) process() {
	for {
		select {
		case <-e.done:
			return
		case batch := <-e.rcv:
			var exportErr error
			for _, span := range batch {
				if err := e.writeSpan(span); err != nil {
					if exportErr == nil {
						exportErr = err
					} else {
						exportErr = fmt.Errorf("%w -- %v", err, exportErr)
					}

				}
				// b, _ := span.MarshalJSON()
				// if _, err := e.w.Write(b); err != nil {
				// }
			}

			if exportErr != nil {
				logx.Error("export errored", attr.String("error", exportErr.Error()))
			}
		}
	}

}

func (e writerExporter) writeSpan(span SpanData) error {
	b, _ := span.MarshalJSON()
	if _, err := e.w.Write(b); err != nil {
		return err
	}
	return nil
}

func (e writerExporter) Shutdown(ctx context.Context) error {
	if wc, ok := e.w.(interface {
		Close() error
	}); ok {
		return wc.Close()
	}
	e.done <- struct{}{}
	return nil
}

func Writer(w io.Writer) Exporter {
	we := writerExporter{
		rcv:  make(chan []SpanData),
		done: make(chan struct{}),
		enc:  jsonEnc{},
		w:    w,
	}
	go we.process()

	return we
}
