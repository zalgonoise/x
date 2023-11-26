package audio

// Registry is a generic type that registers and loads values on a specific type context.
//
// Registries are responsible for handling aggregations and compacting values into one, when Load is called.
type Registry[T any] interface {
	// Register stores the input data in the Registry, returning an error if raised
	Register(T) error
	// Loader implements the Load method, which returns a receive-only channel of items of a given type, usually as
	// part of a Registry features.
	//
	// It is the responsibility of the Loader or the Registry to feed items into the returned channel for consumers to
	// use accordingly, according to the Loader or Registry strategy implemented by the type.
	Loader[T]
	// Closer requires the Shutdown method, allowing a Registry to gracefully shutdown.
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
// A Compactor is an optional, configurable component within a Registry, if applicable.
type Compactor[T any] func([]T) (T, error)
