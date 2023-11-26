package audio

// NoOpExtractor returns an Extractor for a given type, that does not perform any operations on the input values,
// and only returns zero values for a given type.
func NoOpExtractor[T any]() Extractor[T] {
	return Extraction[T](func(header Header, float64s []float64) T {
		return *new(T)
	})
}
