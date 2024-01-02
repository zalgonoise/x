package mapping

import (
	"slices"

	"github.com/zalgonoise/cfg"
)

// Index is a Field type that stores all the values map's keys in a slice, accessible as a public type within Index.
//
// This Field type can also be an ordered Index, where the keys are ordered with a specific logic, if a comparison
// function is provided (in WithIndex) when configuring it.
type Index[K comparable, T any] struct {
	Keys []K

	zero   T
	cmp    func(a, b K) int
	values map[K]T
}

// Get fetches the value in a mapping Field for a given key. If the value does not exist, the Field's
// configured zero value is returned. A boolean value is also returned to highlight whether accessing the key was
// successful or not.
func (i *Index[K, T]) Get(key K) (T, bool) {
	var zero K

	if key == zero || len(i.values) == 0 {
		return i.zero, false
	}

	value, ok := i.values[key]
	if !ok {
		return i.zero, false
	}

	return value, true
}

// Set replaces the value of a certain key in the map, or it adds it if it does not exist. The returned boolean value
// represents whether the key is new in the mapping Field or not.
func (i *Index[K, T]) Set(key K, value T) bool {
	_, ok := i.values[key]

	i.values[key] = value

	if ok {
		return !ok
	}

	i.Keys = append(i.Keys, key)

	if i.cmp != nil {
		slices.SortFunc(i.Keys, i.cmp)
	}

	return ok
}

func NewIndex[K comparable, T any](values map[K]T, opts ...cfg.Option[Config[K, T]]) *Index[K, T] {
	config := cfg.New(opts...)

	idx := make([]K, 0, len(values))
	for key := range values {
		idx = append(idx, key)
	}

	if config.cmpFunc != nil {
		slices.SortFunc(idx, config.cmpFunc)
	}

	return &Index[K, T]{
		Keys: idx,

		cmp:    config.cmpFunc,
		zero:   config.zero,
		values: values,
	}
}
