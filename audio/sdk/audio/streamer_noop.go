package audio

import (
	"context"
	"io"
)

type noOpStreamer struct{}

// Stream implements the Streamer interface
//
// This is a no-op call and has no effect.
func (noOpStreamer) Stream(context.Context, io.Reader, chan<- error) {}

// NoOpStreamer returns a no-op Streamer
func NoOpStreamer() Streamer {
	return noOpStreamer{}
}

type noOpStreamExporter struct{}

// Stream implements the Streamer interface.
//
// This is a no-op call and has no effect.
func (noOpStreamExporter) Stream(context.Context, io.Reader, chan<- error) {}

// Export implements the Exporter interface.
//
// This is a no-op call and the returned error is always nil.
func (noOpStreamExporter) Export(Header, []float64) error { return nil }

// ForceFlush implements the Exporter and StreamCloser interfaces.
//
// This is a no-op call and the returned error is always nil.
func (noOpStreamExporter) ForceFlush() error { return nil }

// Shutdown implements the Exporter, Closer and StreamCloser interfaces.
//
// This is a no-op call and the returned error is always nil.
func (noOpStreamExporter) Shutdown(context.Context) error { return nil }

// NoOpStreamExporter returns a no-op StreamExporter
func NoOpStreamExporter() StreamExporter {
	return noOpStreamExporter{}
}
