package schedule

import (
	"log/slog"
	"time"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/cron/schedule/cronlex"
	"github.com/zalgonoise/x/cron/schedule/resolve"
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

var defaultSchedule = cronlex.Schedule{
	Min:      resolve.Everytime{},
	Hour:     resolve.Everytime{},
	DayMonth: resolve.Everytime{},
	Month:    resolve.Everytime{},
	DayWeek:  resolve.Everytime{},
}

func From(s Scheduler, options ...cfg.Option[SchedulerConfig]) (Scheduler, error) {
	if len(options) == 0 {
		return s, nil
	}

	var (
		config    = cfg.New(options...)
		cronSched cronlex.Schedule
		loc       *time.Location
		tracer    trace.Tracer
		m         Metrics
		logger    *slog.Logger
	)

	if sched, ok := (s).(withTrace); ok {
		s = sched.s
		tracer = sched.tracer
	}

	if sched, ok := (s).(withLogs); ok {
		s = sched.s
		logger = sched.logger
	}

	if sched, ok := (s).(withMetrics); ok {
		s = sched.s
		m = sched.m
	}

	sched, ok := (s).(CronSchedule)
	if !ok {
		return s, ErrInvalidScheduler
	}

	loc = sched.Loc
	cronSched = sched.Schedule

	cron, err := newScheduler(config)
	if err != nil {
		return noOpScheduler{}, err
	}

	if cron.(CronSchedule).Schedule == defaultSchedule && cronSched != defaultSchedule {
		c := cron.(CronSchedule)
		c.Schedule = cronSched
		cron = c
	}

	if cron.(CronSchedule).Loc == time.Local && loc != time.Local {
		c := cron.(CronSchedule)
		c.Loc = loc
		cron = c
	}

	switch {
	case config.metrics != nil:
		cron = schedulerWithMetrics(cron, config.metrics)
	case m != nil:
		cron = schedulerWithMetrics(cron, m)
	}

	switch {
	case config.handler != nil:
		cron = schedulerWithLogs(cron, config.handler)
	case logger != nil:
		cron = schedulerWithLogs(cron, logger.Handler())
	}

	switch {
	case config.handler != nil:
		cron = schedulerWithTrace(cron, config.tracer)
	case tracer != nil:
		cron = schedulerWithTrace(cron, tracer)
	}

	return cron, nil
}
