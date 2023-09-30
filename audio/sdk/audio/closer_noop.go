package audio

import "context"

type noOpCloser struct{}

// ForceFlush implements the StreamCloser interface.
//
// This is a no-op implementation and the returned error is always nil.
func (noOpCloser) ForceFlush() error { return nil }

// Shutdown implements the StreamCloser interface.
//
// This is a no-op implementation and the returned error is always nil.
func (noOpCloser) Shutdown(context.Context) error { return nil }

// NoOpCloser returns a no-op StreamCloser.
func NoOpCloser() StreamCloser {
	return noOpCloser{}
}
