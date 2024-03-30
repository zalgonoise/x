package audio

import (
	"github.com/zalgonoise/x/audio/fft"
)

// Emitter is responsible for pushing the output data of an exporter to the
// appropriate destination or backend.
//
// The values received by the Emitter are final, processed values that have been consumed from
// a byte stream, extracted as audio signal, then registered and compacted before being emitted
// with the appropriate EmitPeaks and EmitSpectrum calls, by each Extractor.
//
// The Emitter doesn't care how the Processor, Exporter, and configured components (Extractor, Registry, Compactor)
// arrive to these values nor how often; merely that these should be the values to register as output,
// or in their appropriate backends.
//
// Emitting a value to a destination or backend should not return any errors. This type is implemented
// using packages external to this audio SDK, and as such the caller is invited to handle those kinds of
// errors separately. For example, if the Emitter is writing some kind of metric as text in a file,
// it's probably best to implement it including a logger for registering errors when writing data into that file.
//
// Note: not all implementations need to use the base Exporter logic, therefore they may skip this
// requirement and dependency.
//
// parent: Exporter
// child:
type Emitter interface {
	// EmitPeaks registers peak audio levels, as received by a Registry in an Exporter.
	EmitPeaks(float64)
	// EmitSpectrum registers the (full) frequency spectrum levels, as received by a Registry in an Exporter.
	EmitSpectrum([]fft.FrequencyPower)
	// Closer requires the Shutdown method, allowing an Emitter to gracefully shutdown.
	Closer
}
