package batchreg

import (
	"time"

	"github.com/zalgonoise/x/audio/sdk/audio"
	"github.com/zalgonoise/x/cfg"
)

type Config[T any] struct {
	flushFrequency time.Duration
	maxBatchSize   int

	reg       audio.Registerer[T]
	compactor audio.Compactor[T]
}

func WithRegistry[T any](reg audio.Registerer[T]) cfg.Option[Config[T]] {
	return cfg.Register(func(config Config[T]) Config[T] {
		config.reg = reg

		return config
	})
}

func WithCompactor[T any](compactor audio.Compactor[T]) cfg.Option[Config[T]] {
	return cfg.Register(func(config Config[T]) Config[T] {
		config.compactor = compactor

		return config
	})
}

func WithBatchSize[T any](size int) cfg.Option[Config[T]] {
	return cfg.Register(func(config Config[T]) Config[T] {
		config.maxBatchSize = size

		return config
	})
}

func WithFlushFrequency[T any](dur time.Duration) cfg.Option[Config[T]] {
	return cfg.Register(func(config Config[T]) Config[T] {
		config.flushFrequency = dur

		return config
	})
}
