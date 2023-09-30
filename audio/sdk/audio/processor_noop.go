package audio

import (
	"context"
	"io"
)

type noOpProcessor struct{}

// Process implements the Processor interface.
//
// This is a no-op call.
func (noOpProcessor) Process(context.Context, io.Reader) {}

// Err implements the Processor interface.
//
// This is a no-op call and the returned channel is always nil
func (noOpProcessor) Err() <-chan error { return nil }

// ForceFlush implements the Processor and StreamCloser interfaces.
//
// This is a no-op call and the returned error is always nil
func (noOpProcessor) ForceFlush() error { return nil }

// Shutdown implements the Processor, Closer and StreamCloser interfaces.
//
// This is a no-op call and the returned error is always nil
func (noOpProcessor) Shutdown(context.Context) error { return nil }

// NoOpProcessor returns a no-op Processor
func NoOpProcessor() Processor {
	return noOpProcessor{}
}
