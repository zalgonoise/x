package audio

// Registerer is a generic type that registers and loads values on a specific type context.
//
// Registries are responsible for handling aggregations and compacting values into one, when Load is called
type Registerer[T any] interface {
	Register(T) error
	Loader[T]
}

type Loader[T any] interface {
	Load() <-chan T
}

type noOpRegisterer[T any] struct{}

func (noOpRegisterer[T]) Register(T) error { return nil }
func (noOpRegisterer[T]) Load() <-chan T   { return nil }

// NoOpLoader returns a no-op Registerer, scoped as a Loader interface
func NoOpLoader[T any]() Loader[T] {
	return noOpRegisterer[T]{}
}

// NoOpRegisterer returns a no-op Registerer
func NoOpRegisterer[T any]() Registerer[T] {
	return noOpRegisterer[T]{}
}
