package audio

import "context"

// Registry is a generic type that registers and loads values on a specific type context.
//
// Registries are responsible for handling aggregations and compacting values into one, when Load is called
type Registry[T any] interface {
	// Register stores the input data in the Registry, returning an error if raised
	Register(T) error
	// Loader implements the Load method, which returns a receive-only channel of items of a given type, usually as
	// part of a Registry features.
	//
	// It is the responsibility of the Loader or the Registry to feed items into the returned channel for consumers to
	// use accordingly, according to the Loader or Registry strategy implemented by the type.
	Loader[T]
	// Closer implements the Shutdown method, allowing a Registry to gracefully shutdown.
	Closer
}

// Loader is a generic type that only emits items of a given type from a channel. This channel is accessible via
// the Load method call, and can then be consumed until closed.
//
// It is the responsibility of the Registry or Loader to feed items into this channel as they are registered, or as
// defined in the Registry / Loader strategy implemented by the Registry or Loader type.
type Loader[T any] interface {
	// Load returns a receive-only channel of items of a given type, usually as part of a Registry features.
	//
	// It is the responsibility of the Loader or the Registry to feed items into the returned channel for consumers to
	// use accordingly, according to the Loader or Registry strategy implemented by the type.
	Load() <-chan T
}

// Compactor is a function that creates a summary of a set of values based on a certain rule (max, average, rate, etc)
// returning one single value of the same type and an error if raised.
//
// It should be perceived as the reduce process in a Map-Reduce strategy.
//
// A Compactor is an optional, configurable component within a Registry, if applicable
type Compactor[T any] func([]T) (T, error)

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

// NoOpLoader returns a no-op Registry, scoped as a Loader interface
func NoOpLoader[T any]() Loader[T] {
	return noOpRegistry[T]{}
}

// NoOpRegisterer returns a no-op Registry
func NoOpRegisterer[T any]() Registry[T] {
	return noOpRegistry[T]{}
}
