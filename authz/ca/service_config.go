package ca

import (
	"log/slog"

	"github.com/zalgonoise/cfg"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	logHandler slog.Handler
	tracer     trace.Tracer

	template []cfg.Option[Template]
}

func WithLogger(logger *slog.Logger) cfg.Option[Config] {
	if logger == nil {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register[Config](func(config Config) Config {
		config.logHandler = logger.Handler()

		return config
	})
}

func WithLogHandler(logger slog.Handler) cfg.Option[Config] {
	if logger == nil {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register[Config](func(config Config) Config {
		config.logHandler = logger

		return config
	})
}

func WithTracer(tracer trace.Tracer) cfg.Option[Config] {
	if tracer == nil {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register[Config](func(config Config) Config {
		config.tracer = tracer

		return config
	})
}

func WithTemplate(opts ...cfg.Option[Template]) cfg.Option[Config] {
	if len(opts) == 0 {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register[Config](func(config Config) Config {
		config.template = opts

		return config
	})
}
