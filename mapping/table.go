package mapping

import (
	"github.com/zalgonoise/cfg"
)

type Table[K comparable, T any] struct {
	zero   T
	values map[K]T
}

func (t Table[K, T]) Get(key K) (T, bool) {
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

func NewTable[K comparable, T any](values map[K]T, opts ...cfg.Option[Config[K, T]]) Table[K, T] {
	config := cfg.New(opts...)

	return Table[K, T]{
		zero:   config.zero,
		values: values,
	}
}
