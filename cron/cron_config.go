package cron

import (
	"log/slog"

	"github.com/zalgonoise/x/cfg"
	"go.opentelemetry.io/otel/trace"
)

type RuntimeConfig struct {
	exec []Executor

	logger  *slog.Logger
	metrics CronMetrics
	tracer  trace.Tracer
}

func WithExecutors(executors ...Executor) cfg.Option[RuntimeConfig] {
	return cfg.Register(func(config RuntimeConfig) RuntimeConfig {
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

func WithCronMetrics(m CronMetrics) cfg.Option[RuntimeConfig] {
	return cfg.Register(func(config RuntimeConfig) RuntimeConfig {
		config.metrics = m

		return config
	})
}

func WithCronLogs(logger *slog.Logger) cfg.Option[RuntimeConfig] {
	return cfg.Register(func(config RuntimeConfig) RuntimeConfig {
		config.logger = logger

		return config
	})
}

func WithCronTrace(tracer trace.Tracer) cfg.Option[RuntimeConfig] {
	return cfg.Register(func(config RuntimeConfig) RuntimeConfig {
		config.tracer = tracer

		return config
	})
}
