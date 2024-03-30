package audio

// Extractor is a generic interface for a type that implements the Extract method, which can return a value from
// parsing an audio chunk.
//
// It should be perceived as the map process in a Map-Reduce strategy.
//
// Its Extract method is intended to be executed multiple times on each audio chunk received from a stream, and it is
// configured with a Collector in mind.
//
// parent: Collector
// child:
type Extractor[T any] interface {
	// Extract will analyze the audio chunk (as a slice of float64 values), optionally referring to the audio header's
	// metadata, to extract a value of a given type.
	//
	// It should be perceived as the map process in a Map-Reduce strategy.
	//
	// Extract method is intended to be executed multiple times on each audio chunk received from a stream, and it is
	// configured with a Collector in mind.
	Extract(Header, []float64) T
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
type Extraction[T any] func(Header, []float64) T

// Extract implements the Extractor interface.
//
// It will call itself as a function, using the input parameters as its arguments.
func (e Extraction[T]) Extract(header Header, data []float64) T {
	return e(header, data)
}

// Threshold evaluates a given input value and determines whether it crosses a certain limit or threshold,
// returning a boolean representing it.
//
// It can be used in an Exporter's sequence, for example after the appropriate Extractor pulled meaningful data
// from the audio stream; where the Threshold function will be signaling whether this data should be exported,
// or just skipped through.
type Threshold[T any] func(T) bool
