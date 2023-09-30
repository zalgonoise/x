package audio

import (
	"context"
	"io"
)

type noOpConsumer struct{}

// Consume implements the Consumer interface.
//
// This is a no-op call and the returned values are both always nil.
func (noOpConsumer) Consume(context.Context) (reader io.Reader, err error) { return nil, nil }

// Shutdown implements the Consumer and Closer interfaces.
//
// This is a no-op call and the returned error is always nil.
func (noOpConsumer) Shutdown(context.Context) error { return nil }

// NoOpConsumer returns a no-op Consumer
func NoOpConsumer() Consumer {
	return noOpConsumer{}
}
