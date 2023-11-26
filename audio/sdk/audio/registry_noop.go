package audio

import "context"

type noOpRegistry[T any] struct{}

// Register implements the Registry interface.
//
// This is a no-op call and the returned error is always nil.
func (noOpRegistry[T]) Register(T) error { return nil }

// Load implements the Registry interface.
//
// This is a no-op call and the returned channel is always nil.
func (noOpRegistry[T]) Load() <-chan T { return nil }

// Shutdown implements the Registry and Closer interfaces.
//
// This is a no-op call and the returned error is always nil.
func (noOpRegistry[T]) Shutdown(context.Context) error { return nil }

// NoOpLoader returns a no-op Registry, scoped as a Loader interface.
func NoOpLoader[T any]() Loader[T] {
	return noOpRegistry[T]{}
}

// NoOpRegistry returns a no-op Registry.
func NoOpRegistry[T any]() Registry[T] {
	return noOpRegistry[T]{}
}
