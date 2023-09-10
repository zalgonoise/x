package cron

import (
	"log/slog"

	"github.com/zalgonoise/x/cfg"
	"github.com/zalgonoise/x/cron/selector"
	"go.opentelemetry.io/otel/trace"
)

const (
	minBufferSize     = 64
	defaultBufferSize = 1024
)

type Config struct {
	sel           selector.Selector
	errBufferSize int

	logger  *slog.Logger
	metrics Metrics
	tracer  trace.Tracer
}

func WithSelector(sel selector.Selector) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.sel = sel

		return config
	})
}

func WithErrorBufferSize(size int) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.errBufferSize = size

		return config
	})
}

func WithMetrics(m Metrics) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.metrics = m

		return config
	})
}

func WithLogs(logger *slog.Logger) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.logger = logger

		return config
	})
}

func WithTrace(tracer trace.Tracer) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.tracer = tracer

		return config
	})
}
