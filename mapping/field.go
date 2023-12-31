package mapping

import (
	"slices"

	"github.com/zalgonoise/cfg"
)

type Field[K comparable, T any] interface {
	Get(key K) (T, bool)
}

func New[K comparable, T any](values map[K]T, opts ...cfg.Option[Config[K, T]]) Field[K, T] {
	config := cfg.New(opts...)

	if !config.indexed {
		return Table[K, T]{
			zero:   config.zero,
			values: values,
		}
	}

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
