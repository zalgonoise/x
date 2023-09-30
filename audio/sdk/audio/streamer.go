package audio

import (
	"context"
	"io"
)

// Streamer describes types that are able to streamer an audio signal
type Streamer interface {
	Stream(ctx context.Context, r io.Reader, errCh chan<- error)
}

type StreamExporter interface {
	Streamer
	Exporter
}

type streamExporter struct {
	Streamer
	Exporter
}

// NewStreamExporter creates a data structure encapsulating two separate Streamer and Exporter interfaces
// as one StreamExporter interface
//
// It is the responsibility of the caller to configure these accordingly, considering that certain (if not most)
// implementations should require a connection or configuration between the two components.
//
// The resulting type is a simple struct embedding both interfaces, allowing access to all StreamExporter methods.
func NewStreamExporter(streamer Streamer, exporter Exporter) StreamExporter {
	switch {
	case streamer == nil && exporter == nil:
		return NoOpStreamExporter()
	case streamer == nil:
		streamer = NoOpStreamer()
	case exporter == nil:
		exporter = NoOpExporter()
	}

	return streamExporter{
		Streamer: streamer,
		Exporter: exporter,
	}
}
