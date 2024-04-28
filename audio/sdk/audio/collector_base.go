package audio

import (
	"context"
	"errors"

	"github.com/zalgonoise/x/audio/encoding/wav"
)

const (
	errAlloc = 2
)

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
func (c collector[T]) Collect(ctx context.Context, header *wav.Header, data []float64) error {
	return c.registry.Register(c.extractor.Extract(ctx, header, data))
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
	errs := make([]error, 0, errAlloc)

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
	switch {
	case extractor == nil && registry == nil:
		return NoOpCollector[T]()
	case extractor == nil:
		extractor = NoOpExtractor[T]()
	case registry == nil:
		registry = NoOpRegistry[T]()
	}

	return collector[T]{
		extractor: extractor,
		registry:  registry,
	}
}
