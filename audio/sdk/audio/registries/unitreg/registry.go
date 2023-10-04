package unitreg

import (
	"context"

	"github.com/zalgonoise/x/audio/sdk/audio"
)

const defaultBufferSize = 64

type unitRegistry[T any] struct {
	ch chan T
}

// Register stores the input data in the audio.Registry, returning an error if raised
//
// This implementation sends the input value to its value channel and returns a nil error.
func (r *unitRegistry[T]) Register(value T) error {
	r.ch <- value

	return nil
}

// Load returns a receive-only channel of items of a given type, usually as part of a Registry features.
//
// The value channel is tied to the Register method, which sends data it receives to this channel; allowing
// asynchronous publishing and consumption of said values.
//
// The audio.Registry's Shutdown method will close this channel.
func (r *unitRegistry[T]) Load() <-chan T {
	return r.ch
}

// Shutdown gracefully stops the audio.Registry.
//
// It effectively closes the value channel returned by the Load method.
func (r *unitRegistry[T]) Shutdown(context.Context) error {
	close(r.ch)

	return nil
}

// New creates a units-audio.Registry for a given type, allocating a buffer with the provided size in the input
// parameter, for its values channel.
//
// If the size of the buffer is a negative number, the default size will be applied.
func New[T any](size int) audio.Registry[T] {
	if size < 0 {
		size = defaultBufferSize
	}

	return &unitRegistry[T]{
		ch: make(chan T, size),
	}
}
