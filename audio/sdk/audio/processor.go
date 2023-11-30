package audio

import (
	"context"
	"io"
)

// Processor is responsible for reading a byte stream and extract useful information from it.
// This can be done with a combination of Collector(s) and Exporter(s).
//
// The Processor will determine the scope of the audio encoding and creates an audio stream from
// the provided io.Reader in its Process method. From that point forward, it consumes the audio from the
// io.Reader in chunks of float64 values that are sent to each configured Exporter.
//
// Processor is not responsible for writing, generating or transforming any data -- it exclusively converts
// the read bytes from the io.Reader into a readable format to be then further processed by any inner modules
// configured in the Processor (Exporter(s), Collector(s), Registry).
//
// Implementations of Processor are responsible for the configuration of these modules, and should allow a modular
// approach to how the caller wants to use the incoming audio. This implies that the Processor pipes all processed audio
// data to its Exporters alike, which in place will run these audio chunks through any configured Compactor(s)
// (if applicable).
type Processor interface {
	// Process reads the byte stream from the input io.Reader and extracts parsed audio, as chunks of float64 values.
	// These chunks are sent to and consumed by any configured Exporter in the Processor
	//
	// The caller should supply a context.Context with a timeout if they desire for the component to run only for a
	// certain amount of time.
	//
	// Process should be called in a goroutine, where the caller checks for the Err error channel, and optionally
	// their own context.Context.Done channel if applicable.
	//
	// Any conversion errors raised during a Process call should stop further reading of the byte stream, and send the
	// error to its error channel, that is exposed via the Err call.
	Process(ctx context.Context, reader io.Reader)
	// Err returns a receiving channel for errors, that allows the caller of a Process method to listen for any raised
	// errors
	Err() <-chan error
	// StreamCloser defines common methods when interacting with a streaming module, targeting actions to either flush
	// the module or to shut it down gracefully.
	StreamCloser
}
