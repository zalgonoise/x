package cfg

// Option describes a generic interface type that can be used to set or define options in (any) configuration data
// structure.
type Option[T any] interface {
	apply(config T) T
}

type configFunc[T any] func(T) T

func (fn configFunc[T]) apply(config T) T {
	return fn(config)
}

// Register creates a new Option for a configuration data structure of type T.
func Register[T any](fn func(T) T) Option[T] {
	return configFunc[T](fn)
}

// New creates a new configuration data structure of type T and applies all
// configuration options passed by the caller; returning the resulting data structure.
func New[T any](options ...Option[T]) T {
	return Set(*new(T), options...)
}

// Set applies all Option configuration options to the input config, of any type. It returns
// a modified version of the input config with all applied options.
func Set[T any](config T, options ...Option[T]) T {
	for i := range options {
		config = options[i].apply(config)
	}

	return config
}