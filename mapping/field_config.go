package mapping

import "github.com/zalgonoise/cfg"

type Config[K comparable, T any] struct {
	zero    T
	indexed bool
	cmpFunc func(a, b K) int
}

// WithZero sets a zero value to be returned whenever a certain key isn't found in the mapping Field.
//
// If this option is not used, the default zero value for the given type will be used instead
// (e.g. "" for strings; 0 for integers; false for booleans; nil for pointers; etc.).
func WithZero[K comparable, T any](zero T) cfg.Option[Config[K, T]] {
	return cfg.Register(func(config Config[K, T]) Config[K, T] {
		config.zero = zero

		return config
	})
}

// WithIndex creates an index of all keys that are accessible via an Index type's keys element. Creating a Field via the
// New function using this option will always return an *Index[K, T] type, and its Index element can be used if cast to
// the concrete type.
//
// Specifying a comparison function (like one such as cmp.Compare, to be used in slices.SortFunc) will order the slice
// of keys in the Index; Otherwise the keys will contain the order as retrieved from the map (random order).
//
// If a Field type is used, the keys slice can also be retrieved using the Keys function.
func WithIndex[T any, K comparable](cmpFunc func(a, b K) int) cfg.Option[Config[K, T]] {
	return cfg.Register(func(config Config[K, T]) Config[K, T] {
		config.indexed = true

		if cmpFunc == nil {
			return config
		}

		config.cmpFunc = cmpFunc

		return config
	})
}
