package audio

import (
	"context"

	"github.com/zalgonoise/x/audio/encoding/wav"
)

// Exporter is responsible for pushing the processed data into a certain destination.
//
// It is of the implementer's responsibility to comply with any requirements needed to export the audio,
// which is out-of-scope for this package.
//
// Exporter will periodically receive data either from the Processor or its Collector functions. The data
// is ready to be pushed to its destination by the time it is passed to the Exporter and there shouldn't be any
// further processing.
//
// Its Export method will propagate the incoming audio data chunk into its Collector(s) (and Compactor(s) if
// configured), or directly to its output target if none are configured.
//
// Exporter also implements StreamCloser as a means to both flush any batched or aggregated values and gracefully
// shutdown the exporter.
//
// parent: Processor
// child: Collector, Emitter
type Exporter interface {
	// Export consumes the audio data chunks from the Processor, preparing them to be pushed to their destination.
	//
	// The Exporter may have a set of Collector configured -- in that case it send the audio data it receives to all
	// Collector. It is also the responsibility of the implementation of Exporter to properly send those aggregations
	// or batches to their correct destination once they're done being collected, or on a frequent flush.
	//
	// The returned error from an Export call is related to an error raised when pushing the values or items to the
	// target, or from any errors raised by the configured Collector types.
	Export(ctx context.Context, header *wav.Header, data []float64) error
	// StreamCloser defines common methods when interacting with a streaming module, targeting actions to either flush
	// the module or to shut it down gracefully.
	StreamCloser
}
