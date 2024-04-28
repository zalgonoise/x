package audio

import (
	"context"

	"github.com/zalgonoise/x/audio/encoding/wav"
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
//
// parent: Exporter
// child: Loader, Extractor, Registry
type Collector[T any] interface {
	// Collect processes audio data by chunks, with its header metadata as reference. It returns an error if raised.
	//
	// Collect involves using the Collector's Extractor to retrieve meaningful data from the signal (of a given
	// data type) and passing this value to its Registry to store, cache or buffer it, for instance.
	Collect(ctx context.Context, header *wav.Header, data []float64) error

	// Loader returns a receive-only channel of a given type, that is used by the Exporter to retrieve processed data
	// from a Collector. Depending on the configured Registry strategy, the Loader will provide data based off of that
	// same Registry.
	Loader[T]

	// StreamCloser allows force-flushing and to gracefully shutting down the Collector.
	StreamCloser
}
