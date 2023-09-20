package audio

import (
	"context"
	"errors"
	"time"
)

// StreamCloser defines common methods when interacting with a streaming module, targeting actions to either flush
// any buffered values or items in the module and to shut it down gracefully.
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

type Closer interface {
	Shutdown(ctx context.Context) error
}

type CloserFunc func(context.Context) error

func (fn CloserFunc) Shutdown(ctx context.Context) error {
	return fn(ctx)
}

type noOpCloser struct{}

func (noOpCloser) ForceFlush() error              { return nil }
func (noOpCloser) Shutdown(context.Context) error { return nil }

// NoOpCloser returns a no-op StreamCloser
func NoOpCloser() StreamCloser {
	return noOpCloser{}
}

func Shutdown(ctx context.Context, timeout time.Duration, closers ...Closer) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if len(closers) == 0 {
		return nil
	}

	errs := make([]error, 0, len(closers))

	for i := range closers {
		if err := closers[i].Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
