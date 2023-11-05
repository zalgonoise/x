package schedule

import (
	"log/slog"
	"time"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/errs"
	"go.opentelemetry.io/otel/trace"
)

const (
	errDomain = errs.Domain("x/cron/schedule")

	ErrInvalid = errs.Kind("invalid")

	ErrScheduler = errs.Entity("scheduler")
)

var ErrInvalidScheduler = errs.WithDomain(errDomain, ErrInvalid, ErrScheduler)

type SchedulerConfig struct {
	cronString string
	loc        *time.Location

	handler slog.Handler
	metrics Metrics
	tracer  trace.Tracer
}

func WithSchedule(cronString string) cfg.Option[SchedulerConfig] {
	if cronString == "" {
		return cfg.NoOp[SchedulerConfig]{}
	}

	return cfg.Register(func(config SchedulerConfig) SchedulerConfig {
		config.cronString = cronString

		return config
	})
}

func WithLocation(loc *time.Location) cfg.Option[SchedulerConfig] {
	if loc == nil {
		return cfg.NoOp[SchedulerConfig]{}
	}

	return cfg.Register(func(config SchedulerConfig) SchedulerConfig {
		config.loc = loc

		return config
	})
}

func WithMetrics(m Metrics) cfg.Option[SchedulerConfig] {
	if m == nil {
		return cfg.NoOp[SchedulerConfig]{}
	}

	return cfg.Register(func(config SchedulerConfig) SchedulerConfig {
		config.metrics = m

		return config
	})
}

func WithLogger(logger *slog.Logger) cfg.Option[SchedulerConfig] {
	if logger == nil {
		return cfg.NoOp[SchedulerConfig]{}
	}

	return cfg.Register(func(config SchedulerConfig) SchedulerConfig {
		config.handler = logger.Handler()

		return config
	})
}

func WithLogHandler(handler slog.Handler) cfg.Option[SchedulerConfig] {
	if handler == nil {
		return cfg.NoOp[SchedulerConfig]{}
	}

	return cfg.Register(func(config SchedulerConfig) SchedulerConfig {
		config.handler = handler

		return config
	})
}

func WithTrace(tracer trace.Tracer) cfg.Option[SchedulerConfig] {
	if tracer == nil {
		return cfg.NoOp[SchedulerConfig]{}
	}

	return cfg.Register(func(config SchedulerConfig) SchedulerConfig {
		config.tracer = tracer

		return config
	})
}
