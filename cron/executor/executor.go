package executor

import (
	"context"
	"errors"
	"time"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/errs"

	"github.com/zalgonoise/x/cron/schedule"
)

const (
	defaultID    = "cron.executor"
	bufferPeriod = 100 * time.Millisecond

	errDomain = errs.Domain("x/cron/executor")

	ErrEmpty = errs.Kind("empty")

	ErrRunnerList = errs.Entity("runners list")
	ErrScheduler  = errs.Entity("scheduler")
)

var (
	ErrEmptyRunnerList = errs.WithDomain(errDomain, ErrEmpty, ErrRunnerList)
	ErrEmptyScheduler  = errs.WithDomain(errDomain, ErrEmpty, ErrScheduler)
)

type Runner interface {
	Run(ctx context.Context) error
}

type Runnable func(ctx context.Context) error

func (r Runnable) Run(ctx context.Context) error {
	if r == nil {
		return nil
	}

	return r(ctx)
}

type Executor interface {
	Exec(ctx context.Context) error
	Next(ctx context.Context) time.Time
	ID() string
}

type Executable struct {
	id      string
	cron    schedule.Scheduler
	runners []Runner
}

func (e Executable) Next(ctx context.Context) time.Time {
	return e.cron.Next(ctx, time.Now())
}

func (e Executable) Exec(ctx context.Context) error {
	execCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	now := time.Now()
	next := e.cron.Next(execCtx, now)
	timer := time.NewTimer(next.Sub(now))
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-timer.C:
			// avoid executing before it's time, as it may trigger repeated runs
			if preTriggerDuration := next.Sub(time.Now()); preTriggerDuration > 0 {
				time.Sleep(preTriggerDuration + bufferPeriod)
			}

			runnerErrs := make([]error, 0, len(e.runners))

			for i := range e.runners {
				if err := e.runners[i].Run(ctx); err != nil {
					runnerErrs = append(runnerErrs, err)
				}
			}

			return errors.Join(runnerErrs...)
		}
	}
}

func (e Executable) ID() string {
	return e.id
}

func New(id string, options ...cfg.Option[Config]) (Executor, error) {
	config := cfg.New(options...)

	exec, err := newExecutable(id, config)
	if err != nil {
		return noOpExecutor{}, err
	}

	if config.metrics != nil {
		exec = executorWithMetrics(exec, config.metrics)
	}

	if config.handler != nil {
		exec = executorWithLogs(exec, config.handler)
	}

	if config.tracer != nil {
		exec = executorWithTrace(exec, config.tracer)
	}

	return exec, nil
}

func newExecutable(id string, config Config) (Executor, error) {
	// validate input
	if id == "" {
		id = defaultID
	}

	if len(config.runners) == 0 {
		return noOpExecutor{}, ErrEmptyRunnerList
	}

	if config.scheduler == nil && config.cronString == "" {
		return noOpExecutor{}, ErrEmptyScheduler
	}

	var sched schedule.Scheduler
	switch {
	case config.scheduler != nil:
		// scheduler is provided, ignore cron string and location
		sched = config.scheduler
	default:
		// create a new scheduler from config
		opts := make([]cfg.Option[schedule.SchedulerConfig], 0, 2)

		if config.cronString != "" {
			opts = append(opts, schedule.WithSchedule(config.cronString))
		}

		if config.loc != nil {
			opts = append(opts, schedule.WithLocation(config.loc))
		}

		var err error

		sched, err = schedule.New(opts...)
		if err != nil {
			return noOpExecutor{}, err
		}
	}

	// return the object with the provided runners
	return Executable{
		id:      id,
		cron:    sched,
		runners: config.runners,
	}, nil
}

func NoOp() Executor {
	return noOpExecutor{}
}

type noOpExecutor struct{}

func (e noOpExecutor) Exec(_ context.Context) error {
	return nil
}

func (e noOpExecutor) Next(_ context.Context) (t time.Time) {
	return t
}

func (e noOpExecutor) ID() string {
	return ""
}
