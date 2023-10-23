package fts

import (
	"log/slog"

	"github.com/zalgonoise/x/cfg"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	uri string

	logHandler slog.Handler
	metrics    Metrics
	tracer     trace.Tracer
}

func WithURI(uri string) cfg.Option[Config] {
	return cfg.Register[Config](func(config Config) Config {
		config.uri = uri

		return config
	})
}

func WithLogger(logger *slog.Logger) cfg.Option[Config] {
	return cfg.Register[Config](func(config Config) Config {
		config.logHandler = logger.Handler()

		return config
	})
}

func WithLogHandler(handler slog.Handler) cfg.Option[Config] {
	return cfg.Register[Config](func(config Config) Config {
		config.logHandler = handler

		return config
	})
}

func WithMetrics(metrics Metrics) cfg.Option[Config] {
	return cfg.Register[Config](func(config Config) Config {
		config.metrics = metrics

		return config
	})
}

func WithTracer(tracer trace.Tracer) cfg.Option[Config] {
	return cfg.Register[Config](func(config Config) Config {
		config.tracer = tracer

		return config
	})
}
