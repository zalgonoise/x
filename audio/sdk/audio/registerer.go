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
