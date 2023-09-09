package cron

import (
	"log/slog"

	"github.com/zalgonoise/x/cfg"
	"github.com/zalgonoise/x/cron/executor"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	exec []executor.Executor

	logger  *slog.Logger
	metrics CronMetrics
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

func WithCronMetrics(m CronMetrics) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.metrics = m

		return config
	})
}

func WithCronLogs(logger *slog.Logger) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.logger = logger

		return config
	})
}

func WithCronTrace(tracer trace.Tracer) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.tracer = tracer

		return config
	})
}
