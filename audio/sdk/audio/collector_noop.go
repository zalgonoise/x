package audio

import (
	"context"

	"github.com/zalgonoise/x/audio/encoding/wav"
)

type noOpCollector[T any] struct{}

// Collect implements the Collector interface.
//
// This is a no-op call and the returned error is always nil.
func (noOpCollector[T]) Collect(*wav.Header, []float64) error { return nil }

// Load implements the Collector and Loader interfaces.
//
// This is a no-op call and the returned error is always nil.
func (noOpCollector[T]) Load() <-chan T { return nil }

// ForceFlush implements the Collector and StreamCloser interfaces.
//
// This is a no-op call and the returned error is always nil.
func (noOpCollector[T]) ForceFlush() error { return nil }

// Shutdown implements the Collector, Closer and StreamCloser interfaces.
//
// This is a no-op call and the returned error is always nil.
func (noOpCollector[T]) Shutdown(context.Context) error { return nil }

// NoOpCollector returns a no-op Collector for a given type.
func NoOpCollector[T any]() Collector[T] {
	return noOpCollector[T]{}
}
