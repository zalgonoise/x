package mapping

import (
	"github.com/zalgonoise/cfg"
)

// Table is a simple Field type, that provides control over the returned zero value when accessing keys in a map.
type Table[K comparable, T any] struct {
	zero   T
	values map[K]T
}

// Get fetches the value in a mapping Field for a given key. If the value does not exist, the Field's
// configured zero value is returned. A boolean value is also returned to highlight whether accessing the key was
// successful or not.
func (t *Table[K, T]) Get(key K) (T, bool) {
	var zero K

	if key == zero || len(t.values) == 0 {
		return t.zero, false
	}

	value, ok := t.values[key]
	if !ok {
		return t.zero, false
	}

	return value, true
}

// Set replaces the value of a certain key in the map, or it adds it if it does not exist. The returned boolean value
// represents whether the key is new in the mapping Field or not.
func (t *Table[K, T]) Set(key K, setter Setter[T]) bool {
	value, exists := t.values[key]

	if !exists {
		value = *new(T)

		t.values[key] = value
	}

	newValue, added := setter(value)

	t.values[key] = newValue

	return added
}

func NewTable[K comparable, T any](values map[K]T, opts ...cfg.Option[Config[K, T]]) *Table[K, T] {
	config := cfg.New(opts...)

	return &Table[K, T]{
		zero:   config.zero,
		values: values,
	}
}
