package cron

import (
	"log/slog"
	"time"

	"github.com/zalgonoise/x/cfg"
	"github.com/zalgonoise/x/cron/schedule"
	"go.opentelemetry.io/otel/trace"
)

type ExecutorConfig struct {
	scheduler  schedule.Scheduler
	cronString string
	loc        *time.Location
	runners    []Runner

	logger  *slog.Logger
	metrics ExecutorMetrics
	tracer  trace.Tracer
}

func WithRunners(runners ...Runner) cfg.Option[ExecutorConfig] {
	return cfg.Register(func(config ExecutorConfig) ExecutorConfig {
		if len(runners) == 0 {
			return config
		}

		if len(config.runners) == 0 {
			config.runners = runners

			return config
		}

		config.runners = append(config.runners, runners...)

		return config
	})
}

func WithScheduler(sched schedule.Scheduler) cfg.Option[ExecutorConfig] {
	return cfg.Register(func(config ExecutorConfig) ExecutorConfig {
		config.scheduler = sched

		return config
	})
}

func WithSchedule(cronString string) cfg.Option[ExecutorConfig] {
	return cfg.Register(func(config ExecutorConfig) ExecutorConfig {
		config.cronString = cronString

		return config
	})
}

func WithLocation(loc *time.Location) cfg.Option[ExecutorConfig] {
	return cfg.Register(func(config ExecutorConfig) ExecutorConfig {
		config.loc = loc

		return config
	})
}

func WithExecutorMetrics(m ExecutorMetrics) cfg.Option[ExecutorConfig] {
	return cfg.Register(func(config ExecutorConfig) ExecutorConfig {
		config.metrics = m

		return config
	})
}

func WithExecutorLogs(logger *slog.Logger) cfg.Option[ExecutorConfig] {
	return cfg.Register(func(config ExecutorConfig) ExecutorConfig {
		config.logger = logger

		return config
	})
}

func WithExecutorTrace(tracer trace.Tracer) cfg.Option[ExecutorConfig] {
	return cfg.Register(func(config ExecutorConfig) ExecutorConfig {
		config.tracer = tracer

		return config
	})
}
