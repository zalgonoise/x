package audio

import (
	"context"
	"io"
)

// Consumer is responsible for consuming, extracting, reading an audio source and returning a byte streamer to read it,
// as an io.Reader.
//
// Implementations should be able to create and return the io.Reader in a non-blocking way, to be read by a Processor.
// Ideally, the implementations of Consumer will not read an entire audio streamer before returning the io.Reader.
//
// Be whatever the audio source it may be, the byte streamer should be valid audio encoded in a supported format. This is
// the sole responsibility of the Consumer -- anything regarding the actual content of the byte streamer would fall on a
// Processor's scope.
//
// parent:
// child: Processor
type Consumer interface {
	// Consume interacts with the audio source to extract its audio content or streamer as an io.Reader.
	//
	// Implementers of the Consumer must return a byte streamer that can be decoded with a supported encoding format.
	//
	// The origin of the byte streamer can be anything, such an os.File, http.Response.Body, bytes.Buffer, etc. The origin
	// of the target audio and any requirements to Consume it are of the Consumers responsibility, and are prepared
	// and / or verified by the time a Consumer is created.
	//
	// The error returned by a Consume call must point to any issues raised in the process of preparing or extracting the
	// returned io.Reader. It must not be related to any of the content of the byte streamer in the io.Reader.
	Consume(ctx context.Context) (reader io.Reader, err error)
	// Shutdown gracefully shuts down the Consumer.
	//
	// If the audio source has an open connection with the Consumer, it is the responsibility of the Shutdown call to
	// close it. For example, if the audio source is an os.File or the body of a http.Response, it is best to close the
	// readers through their io.Closer implementation.
	//
	// The caller is responsible for applying any desired timeout to the Shutdown call. Implementations of Consumer are
	// responsible for imposing any defaults for the same timeouts.
	//
	// The returned error points to any issue or issues raised during this process. If possible, the shutdown process
	// should still continue and close the Consumer on this call.
	Shutdown(ctx context.Context) error
}
