package audio

import (
	"context"
	"io"
)

// Streamer describes types that are able to streamer an audio signal
//
// The Streamer interface is implemented by audio encoder packages as a means of reading
// from an audio source and converting its signal as audio float values.
//
// Streamer types should define Stream as a blocking call to be issued asynchronously (in a goroutine),
// therefore its io.Reader and error channel parameters.
type Streamer interface {
	Stream(ctx context.Context, r io.Reader, errCh chan<- error)
}

// StreamExporter describes a type that implements both Streamer and Exporter interfaces.
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
// The resulting type is a simple struct embedding both interfaces, allowing access to all StreamExporter methods
// with no alterations or method overloading involved.
//
// parent: wav.Stream
// child:
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
