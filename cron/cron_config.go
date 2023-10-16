package cron

import (
	"log/slog"

	"github.com/zalgonoise/x/cfg"
	"go.opentelemetry.io/otel/trace"
)

const (
	minBufferSize     = 64
	defaultBufferSize = 1024
)

type Config struct {
	errBufferSize int

	handler slog.Handler
	metrics Metrics
	tracer  trace.Tracer
}

func WithErrorBufferSize(size int) cfg.Option[Config] {
	if size < 0 {
		size = defaultBufferSize
	}

	return cfg.Register(func(config Config) Config {
		config.errBufferSize = size

		return config
	})
}

func WithMetrics(m Metrics) cfg.Option[Config] {
	if m == nil {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(config Config) Config {
		config.metrics = m

		return config
	})
}

func WithLogger(logger *slog.Logger) cfg.Option[Config] {
	if logger == nil {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(config Config) Config {
		config.handler = logger.Handler()

		return config
	})
}

func WithLogHandler(handler slog.Handler) cfg.Option[Config] {
	if handler == nil {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(config Config) Config {
		config.handler = handler

		return config
	})
}

func WithTrace(tracer trace.Tracer) cfg.Option[Config] {
	if tracer == nil {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(config Config) Config {
		config.tracer = tracer

		return config
	})
}
