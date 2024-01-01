package mapping

import (
	"slices"

	"github.com/zalgonoise/cfg"
)

type Index[K comparable, T any] struct {
	Keys []K

	zero   T
	values map[K]T
}

func (i Index[K, T]) Get(key K) (T, bool) {
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

func NewIndex[K comparable, T any](values map[K]T, opts ...cfg.Option[Config[K, T]]) Index[K, T] {
	config := cfg.New(opts...)

	idx := make([]K, 0, len(values))
	for key := range values {
		idx = append(idx, key)
	}

	if config.cmpFunc != nil {
		slices.SortFunc(idx, config.cmpFunc)
	}

	return Index[K, T]{
		Keys: idx,

		zero:   config.zero,
		values: values,
	}
}
