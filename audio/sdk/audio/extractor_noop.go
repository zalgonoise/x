package audio

import (
	"context"

	"github.com/zalgonoise/x/audio/encoding/wav"
)

// NoOpExtractor returns an Extractor for a given type, that does not perform any operations on the input values,
// and only returns zero values for a given type.
func NoOpExtractor[T any]() Extractor[T] {
	return Extraction[T](func(ctx context.Context, header *wav.Header, float64s []float64) T {
		return *new(T)
	})
}

// NoOpThreshold returns a Threshold for a given type, that does not perform any operations on the input values,
// and always returns a true value when called (representing the absence of a threshold).
func NoOpThreshold[T any]() Threshold[T] {
	return func(T) bool {
		return true
	}
}
