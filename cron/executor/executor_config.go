package executor

import (
	"log/slog"
	"time"

	"github.com/zalgonoise/cfg"
	"go.opentelemetry.io/otel/trace"

	"github.com/zalgonoise/x/cron/schedule"
)

type Config struct {
	scheduler  schedule.Scheduler
	cronString string
	loc        *time.Location

	runners []Runner

	handler slog.Handler
	metrics ExecutorMetrics
	tracer  trace.Tracer
}

func WithRunners(runners ...Runner) cfg.Option[Config] {
	if len(runners) == 0 {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(config Config) Config {
		if len(config.runners) == 0 {
			config.runners = runners

			return config
		}

		config.runners = append(config.runners, runners...)

		return config
	})
}

func WithScheduler(sched schedule.Scheduler) cfg.Option[Config] {
	if sched == nil {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(config Config) Config {
		config.scheduler = sched

		return config
	})
}

func WithSchedule(cronString string) cfg.Option[Config] {
	if cronString == "" {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(config Config) Config {
		config.cronString = cronString

		return config
	})
}

func WithLocation(loc *time.Location) cfg.Option[Config] {
	if loc == nil {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(config Config) Config {
		config.loc = loc

		return config
	})
}

func WithMetrics(m ExecutorMetrics) cfg.Option[Config] {
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
