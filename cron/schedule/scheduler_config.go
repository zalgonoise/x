package schedule

import (
	"log/slog"
	"time"

	"github.com/zalgonoise/parse"
	"github.com/zalgonoise/x/cfg"
	"go.opentelemetry.io/otel/trace"
)

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

func From(s Scheduler, options ...cfg.Option[SchedulerConfig]) (Scheduler, error) {
	if len(options) == 0 {
		return s, nil
	}

	var (
		config = cfg.New(options...)
		out    CronSchedule
	)

	sched, ok := (s).(CronSchedule)
	if !ok {
		return s, ErrInvalidScheduler
	}

	if config.cronString != "" {
		cron, err := parse.Run([]byte(config.cronString), initState, initParse, process)
		if err != nil {
			return s, err
		}

		out = cron

		if sched.Loc != nil {
			out.Loc = sched.Loc
		}
	}

	if config.loc != nil {
		if out.Loc == nil {
			out = sched
		}

		out.Loc = config.loc
	}

	return out, nil
}
