package batchreg

import (
	"time"

	"github.com/zalgonoise/x/audio/sdk/audio"
	"github.com/zalgonoise/x/audio/sdk/audio/compactors"
	"github.com/zalgonoise/x/audio/sdk/audio/registries/unitreg"
	"github.com/zalgonoise/x/cfg"
)

const (
	defaultFlushFrequency = 500 * time.Millisecond
	defaultMaxBatchSize   = 256
	minimumBatchSize      = 64
)

func defaultConfig[T any]() Config[T] {
	return Config[T]{
		flushFrequency: defaultFlushFrequency,
		maxBatchSize:   defaultMaxBatchSize,
		compactor:      compactors.Last[T],
		reg:            unitreg.New[T](defaultMaxBatchSize),
	}
}

type Config[T any] struct {
	flushFrequency time.Duration
	maxBatchSize   int

	reg       audio.Registry[T]
	compactor audio.Compactor[T]
}

func WithRegistry[T any](reg audio.Registry[T]) cfg.Option[Config[T]] {
	if reg == nil {
		return cfg.NoOp[Config[T]]{}
	}

	return cfg.Register(func(config Config[T]) Config[T] {
		config.reg = reg

		return config
	})
}

func WithCompactor[T any](compactor audio.Compactor[T]) cfg.Option[Config[T]] {
	if compactor == nil {
		return cfg.NoOp[Config[T]]{}
	}

	return cfg.Register(func(config Config[T]) Config[T] {
		config.compactor = compactor

		return config
	})
}

func WithBatchSize[T any](size int) cfg.Option[Config[T]] {
	if size < minimumBatchSize {
		return cfg.NoOp[Config[T]]{}
	}

	return cfg.Register(func(config Config[T]) Config[T] {
		config.maxBatchSize = size

		return config
	})
}

func WithFlushFrequency[T any](dur time.Duration) cfg.Option[Config[T]] {
	if dur == 0 {
		return cfg.NoOp[Config[T]]{}
	}

	return cfg.Register(func(config Config[T]) Config[T] {
		config.flushFrequency = dur

		return config
	})
}
