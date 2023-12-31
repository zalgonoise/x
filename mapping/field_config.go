package mapping

import "github.com/zalgonoise/cfg"

type Config[K comparable, T any] struct {
	zero    T
	indexed bool
	cmpFunc func(a K, b K) int
}

func WithZero[K comparable, T any](zero T) cfg.Option[Config[K, T]] {
	return cfg.Register(func(config Config[K, T]) Config[K, T] {
		config.zero = zero

		return config
	})
}

func WithIndex[K comparable, T any](cmpFunc func(a K, b K) int) cfg.Option[Config[K, T]] {
	return cfg.Register(func(config Config[K, T]) Config[K, T] {
		config.indexed = true

		if cmpFunc == nil {
			return config
		}

		config.cmpFunc = cmpFunc

		return config
	})
}
