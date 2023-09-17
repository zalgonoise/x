package audio

import (
	"github.com/zalgonoise/x/audio/wav/header"
)

// Collector is a generic type that is able to parse incoming audio chunks to retrieve
// meaningful information about the signal.
//
// A Collector can process the audio data and extract whatever it wants, creating suitable
// metrics for its needs.
//
// It is the responsibility of the Exporter to store collected values and push them to the
// appropriate backend
type Collector[T any] interface {
	Collect(h *header.Header, data []float64) error
	ForceFlush() error

	Loader[T]
}

type collector[T any] struct {
	extractor  Extractor[T]
	registerer Registerer[T]
}

// Collect implements the Collector interface.
//
// It will use its inner Registerer and Extractor to register the extracted value from the input.
func (c collector[T]) Collect(h *header.Header, data []float64) error {
	return c.registerer.Register(c.extractor.Extract(h, data))
}

func (c collector[T]) Load() <-chan T {
	return c.registerer.Load()
}

func (c collector[T]) ForceFlush() error {
	if flusher, ok := c.registerer.(interface {
		ForceFlush() error
	}); ok {
		if err := flusher.ForceFlush(); err != nil {
			return err
		}
	}

	return nil
}

// NewCollector creates a Collector from hte input Extractor and Registerer
func NewCollector[T any](extractor Extractor[T], registerer Registerer[T]) Collector[T] {
	if extractor == nil || registerer == nil {
		return nil
	}

	return collector[T]{
		extractor:  extractor,
		registerer: registerer,
	}
}

// Compactor is a function that creates a summary of a set of values based on a certain rule (max, average, rate, etc)
// returning one single value of the same type and an error if raised.
type Compactor[T any] func([]T) (T, error)

type noOpCollector[T any] struct{}

func (noOpCollector[T]) Collect(*header.Header, []float64) error {
	return nil
}

func (noOpCollector[T]) ForceFlush() error {
	return nil
}

func (noOpCollector[T]) Load() <-chan T {
	return nil
}

func NoOpCollector[T any]() Collector[T] {
	return noOpCollector[T]{}
}
