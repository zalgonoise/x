package audio

import "context"

// StreamCloser defines common methods when interacting with a streaming module, targeting actions to either flush
// the module or to shut it down gracefully.
type StreamCloser interface {
	// ForceFlush is used in implementations that buffer or batch their values or items, as a means of immediately
	// exporting any values or items that are in-memory to the next destination or component.
	//
	// ForceFlush should be a blocking function which returns only when all the buffered values or items are completely
	// flushed. It is recommended, however, that implementations set their own default timeouts; optionally configurable.
	//
	// The returned error should point to an issue raised when pushing the values or items to the next destination or
	// component, or if ForceClose exits due to a timeout.
	ForceFlush() error
	// Shutdown gracefully shuts down the component.
	//
	// It is the responsibility of the Shutdown call to flush any buffered values or items if they exist, and to close any
	// open connections.
	//
	// The caller is responsible for applying any desired timeout to the Shutdown call. Implementations of StreamCloser
	// are responsible for imposing any defaults for the same timeouts.
	//
	// The returned error points to any issue or issues raised during this process. If possible, the shutdown process
	// should still continue and close the Consumer on this call.
	Shutdown(ctx context.Context) error
}
