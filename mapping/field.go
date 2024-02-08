package mapping

import (
	"sync"

	"github.com/zalgonoise/cfg"
)

// Field describes the capabilities of a dynamic field mapping type, that is to Get and Set values of a certain type,
// using comparable key types.
//
// The Get method fetches the value in a mapping Field for a given key. If the value does not exist, the Field's
// configured zero value is returned. A boolean value is also returned to highlight whether accessing the key was
// successful or not.
//
// Set replaces the value of a certain key in the map, or it adds it if it does not exist. The returned boolean value
// represents whether the key is new in the mapping Field or not.
//
// Implementations of Field include Table (a set of key-value pairs with a configurable zero value), Index (like a Table
// but with the ability to index ordered or unordered keys) and SyncField -- the latter being a synchronized
// implementation of either a Table or Index Field.
type Field[K comparable, T any] interface {
	// Get fetches the value in a mapping Field for a given key. If the value does not exist, the Field's
	// configured zero value is returned. A boolean value is also returned to highlight whether accessing the key was
	// successful or not.
	Get(key K) (T, bool)
	// Set replaces the value of a certain key in the map, or it adds it if it does not exist. The returned boolean value
	// represents whether the key is new in the mapping Field or not.
	Set(key K, setter Setter[T]) bool
}

// Setter is a generic function type that applies a new value in replacement of a former value of type T. It should
// return the new (or same) T value and a boolean representing if the item was newly set (from a zero value) or added;
// where a false value represents a substitution.
type Setter[T any] func(old T) (T, bool)

// New creates a Field type appropriate to the configured options (either a *Table[K, T] type, or an *Index[K, T] type.
//
// Both implementations can be of a SyncField type, if the WithMutex option is used.
func New[K comparable, T any](values map[K]T, opts ...cfg.Option[Config[K, T]]) Field[K, T] {
	config := cfg.New(opts...)
	field := newField(values, opts...)

	if config.synced {
		return SyncField[K, T]{
			field: field,
			mu:    &sync.RWMutex{},
		}
	}

	return field
}

func newField[K comparable, T any](values map[K]T, opts ...cfg.Option[Config[K, T]]) Field[K, T] {
	config := cfg.New(opts...)

	if !config.indexed {
		return NewTable(values, opts...)
	}

	return NewIndex(values, opts...)
}

// Keys provides access to all keys in a Field, regardless if it's a *Table[K, T] type, or an *Index[K, T] type.
func Keys[K comparable, T any](field Field[K, T]) []K {
	switch f := field.(type) {
	case SyncField[K, T]:
		return Keys(f.field)
	case *Index[K, T]:
		return f.Keys
	case *Table[K, T]:
		keys := make([]K, 0, len(f.values))

		for k := range f.values {
			keys = append(keys, k)
		}

		return keys
	default:
		return nil
	}
}
