package cfg

// Option describes a generic interface type that can be used to set or define options in (any) configuration data
// structure.
type Option[T any] interface {
	apply(config T) T
}

// OptionFunc is a function type which implements the Option interface.
type OptionFunc[T any] func(T) T

func (fn OptionFunc[T]) apply(config T) T {
	return fn(config)
}

// Register creates a new Option for a configuration data structure of type T.
//
// It simply sets the input function as a OptionFunc type, if it isn't nil.
func Register[T any](fn func(T) T) Option[T] {
	if fn == nil {
		return NoOp[T]{}
	}

	return OptionFunc[T](fn)
}

// New creates a new configuration data structure of type T and applies all
// configuration options passed by the caller; returning the resulting data structure.
func New[T any](options ...Option[T]) T {
	return Set(*new(T), options...)
}

// Set applies all Option configuration options to the input config, of any type. It returns
// a modified version of the input config with all applied options.
func Set[T any](config T, options ...Option[T]) T {
	options = nonNil(options...)

	for i := range options {
		config = options[i].apply(config)
	}

	return config
}

func nonNil[T any](options ...Option[T]) []Option[T] {
	opts := make([]Option[T], 0, len(options))

	for i := range options {
		if options[i] == nil {
			continue
		}

		opts = append(opts, options[i])
	}

	return opts
}
