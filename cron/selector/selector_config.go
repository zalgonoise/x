package selector

import (
	"log/slog"

	"github.com/zalgonoise/x/cfg"
	"github.com/zalgonoise/x/cron/executor"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	exec []executor.Executor

	logger  *slog.Logger
	metrics Metrics
	tracer  trace.Tracer
}

func WithExecutors(executors ...executor.Executor) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		if len(executors) == 0 {
			return config
		}

		if len(config.exec) == 0 {
			config.exec = executors

			return config
		}

		config.exec = append(config.exec, executors...)

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
