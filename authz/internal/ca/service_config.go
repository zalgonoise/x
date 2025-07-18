package ca

import (
	"log/slog"

	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/zalgonoise/cfg"

	"github.com/zalgonoise/x/authz/internal/certs"
	"github.com/zalgonoise/x/authz/internal/log"
	"github.com/zalgonoise/x/authz/internal/metrics"
)

type Config struct {
	metrics    Metrics
	logHandler slog.Handler
	tracer     trace.Tracer

	template []cfg.Option[certs.Template]
}

func defaultConfig() Config {
	return Config{
		metrics:    metrics.NoOp(),
		logHandler: log.NoOp().Handler(),
		tracer:     noop.NewTracerProvider().Tracer("ca"),
	}
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

func WithMetrics(m Metrics) cfg.Option[Config] {
	if m == nil {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register[Config](func(config Config) Config {
		config.metrics = m

		return config
	})
}

func WithTemplate(opts ...cfg.Option[certs.Template]) cfg.Option[Config] {
	if len(opts) == 0 {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register[Config](func(config Config) Config {
		config.template = opts

		return config
	})
}
