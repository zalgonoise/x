package audio

import (
	"context"
	"errors"

	"github.com/zalgonoise/x/audio/wav/header"
)

// Collector is a generic type that is able to parse incoming audio chunks to retrieve
// meaningful information about the signal.
//
// A Collector can process the audio data and extract whatever it wants, and is able to
// supply these values (for an exporter to consume) via a channel, in its Loader implementation.
//
// The Collector types are configurable with an Extractor and a Registry, allowing very modular
// configurations to both retrieve values of different types, but to also store / cache / buffer
// them with different strategies.
//
// It is the responsibility of the Exporter to store collected values emitted by a Collector's Loader
// and push them to the appropriate backend.
type Collector[T any] interface {
	// Collect processes audio data by chunks, with its header metadata as reference. It returns an error if raised.
	//
	// Collect involves using the Collector's Extractor to retrieve meaningful data from the signal (of a given
	// data type) and passing this value to its Registry to store, cache or buffer it, for instance.
	Collect(h *header.Header, data []float64) error

	// Loader returns a receive-only channel of a given type, that is used by the Exporter to retrieve processed data
	// from a Collector. Depending on the configured Registry strategy, the Loader will provide data based off of that
	// same Registry.
	Loader[T]

	// StreamCloser allows force-flushing and to gracefully shutting down the Collector.
	StreamCloser
}

type collector[T any] struct {
	extractor Extractor[T]
	registry  Registry[T]
}

// Collect implements the Collector interface.
//
// It processes audio data by chunks, with its header metadata as reference. It returns an error if raised.
//
// Collect involves using the Collector's Extractor to retrieve meaningful data from the signal (of a given
// data type) and passing this value to its Registry to store, cache or buffer it, for instance.
func (c collector[T]) Collect(h *header.Header, data []float64) error {
	return c.registry.Register(c.extractor.Extract(h, data))
}

// Load returns a receive-only channel of a given type, that is used by the Exporter to retrieve processed data
// from a Collector. Depending on the configured Registry strategy, the Loader will provide data based off of that
// same Registry.
func (c collector[T]) Load() <-chan T {
	return c.registry.Load()
}

// ForceFlush flushes any values or items that are pending or cached in the Registry, calling its ForceFlush method
// if it exists.
func (c collector[T]) ForceFlush() error {
	if flusher, ok := c.registry.(interface {
		ForceFlush() error
	}); ok {
		if err := flusher.ForceFlush(); err != nil {
			return err
		}
	}

	return nil
}

// Shutdown gracefully shuts down the component, by calling its Registry and Extractor's Shutdown methods, if they
// exist, returning any errors raised in the process.
func (c collector[T]) Shutdown(ctx context.Context) error {
	errs := make([]error, 0, 2)

	if closer, ok := c.extractor.(interface {
		Shutdown(ctx context.Context) error
	}); ok {
		if err := closer.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if err := c.registry.Shutdown(ctx); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// NewCollector creates a Collector from the input Extractor and Registry.
func NewCollector[T any](extractor Extractor[T], registry Registry[T]) Collector[T] {
	if extractor == nil || registry == nil {
		return nil
	}

	return collector[T]{
		extractor: extractor,
		registry:  registry,
	}
}

type noOpCollector[T any] struct{}

// Collect implements the Collector interface
//
// This is a no-op call and the returned error is always nil
func (noOpCollector[T]) Collect(*header.Header, []float64) error { return nil }

// Load implements the Collector and Loader interfaces
//
// This is a no-op call and the returned error is always nil
func (noOpCollector[T]) Load() <-chan T { return nil }

// ForceFlush implements the Collector and StreamCloser interfaces
//
// This is a no-op call and the returned error is always nil
func (noOpCollector[T]) ForceFlush() error { return nil }

// Shutdown implements the Collector, Closer and StreamCloser interfaces
//
// This is a no-op call and the returned error is always nil
func (noOpCollector[T]) Shutdown(context.Context) error { return nil }

// NoOpCollector returns a no-op Collector for a given type
func NoOpCollector[T any]() Collector[T] {
	return noOpCollector[T]{}
}
