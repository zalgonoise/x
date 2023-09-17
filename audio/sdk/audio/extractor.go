package audio

import (
	"github.com/zalgonoise/x/audio/wav/header"
)

// Extractor is a generic interface for a type that implements the Extract method, which can return a value from
// parsing an audio chunk.
type Extractor[T any] interface {
	Extract(*header.Header, []float64) T
}

// Extraction is a generic function type that serves as an audio processor function,
// but returns any type desired, as appropriate to the analysis, processing, recording, whatever it may be.
//
// It is of the responsibility of the Exporter to position the configured Extractor to actually export the
// aggregations.
//
// The sole responsibility of an Extractor is to convert raw audio (as chunks of float64 values) into anything
// meaningful, that is exported / handled separately. Not all Exporter will need one or more Extractor, however
// these are supposed to be perceived as preset building blocks to work with the incoming audio chunks.
type Extraction[T any] func(*header.Header, []float64) T

func (e Extraction[T]) Extract(h *header.Header, data []float64) T {
	return e(h, data)
}
