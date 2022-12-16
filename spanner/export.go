package spanner

import (
	"bytes"
	"context"
	"io"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/logx"
)

// Exporter is a module that pushes the span data to an output
type Exporter interface {
	// Export pushes the input SpanData `spans` to its output, as a non-blocking
	// function
	Export(ctx context.Context, spans []SpanData) error
	// Shutdown gracefully terminates the Exporter
	Shutdown(ctx context.Context) error
}

type noOpExporter struct{}

// Export pushes the input SpanData `spans` to its output, as a non-blocking
// function
func (noOpExporter) Export(ctx context.Context, spans []SpanData) error {
	return nil
}

// Shutdown gracefully terminates the Exporter
func (noOpExporter) Shutdown(ctx context.Context) error {
	return nil
}

var (
	writerBuffer  = new(bytes.Buffer)
	encoderBuffer = []byte{}
)

type writerExporter struct {
	rcv  chan []SpanData
	done chan struct{}
	w    io.Writer
}

// Export pushes the input SpanData `spans` to its output, as a non-blocking
// function
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
			if err := e.writeSpans(batch); err != nil {
				logx.Error("export errored", attr.String("error", err.Error()))
			}
		}
	}
}

func (e writerExporter) writeSpans(batch []SpanData) (err error) {
	defer writerBuffer.Reset()
	for _, span := range batch {
		encoderBuffer, _ = span.MarshalJSON()
		writerBuffer.Write(encoderBuffer)
	}
	if _, err = e.w.Write(writerBuffer.Bytes()); err != nil {
		return err
	}
	return nil
}

// Shutdown gracefully terminates the Exporter
func (e writerExporter) Shutdown(ctx context.Context) error {
	e.done <- struct{}{}
	if wc, ok := e.w.(interface {
		Close() error
	}); ok {
		return wc.Close()
	}
	return nil
}

// Writer returns an Exporter that is configured to write SpanData as
// text (JSON) to an io.Writer
func Writer(w io.Writer) Exporter {
	we := writerExporter{
		rcv:  make(chan []SpanData),
		done: make(chan struct{}),
		w:    w,
	}
	go we.process()

	return we
}
