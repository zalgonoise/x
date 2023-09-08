package cron

import (
	"context"
	"errors"
	"time"

	"github.com/zalgonoise/x/cfg"
	"github.com/zalgonoise/x/cron/schedule"
	"github.com/zalgonoise/x/errs"
)

const (
	errDomain = errs.Domain("x/cron")

	ErrEmpty = errs.Kind("empty")

	ErrRunnerList     = errs.Entity("runners list")
	ErrScheduler      = errs.Entity("scheduler")
	ErrExecutableList = errs.Entity("executable list")
)

var (
	ErrEmptyRunnerList     = errs.New(errDomain, ErrEmpty, ErrRunnerList)
	ErrEmptyScheduler      = errs.New(errDomain, ErrEmpty, ErrScheduler)
	ErrEmptyExecutableList = errs.New(errDomain, ErrEmpty, ErrExecutableList)
)

type Runner interface {
	Run(ctx context.Context) error
}

type Runnable func(ctx context.Context) error

func (r Runnable) Run(ctx context.Context) error {
	return r(ctx)
}

// TODO: probably better to move the Executor type to its own package
type Executor interface {
	Exec(ctx context.Context) error
	Next(ctx context.Context) time.Time
}

type Executable struct {
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

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-execCtx.Done():
			return execCtx.Err()

		case <-timer.C:
			// avoid executing before it's time, as it may trigger repeated runs
			if pretriggerDuration := next.Sub(time.Now()); pretriggerDuration > 0 {
				time.Sleep(pretriggerDuration + 100*time.Millisecond)
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

func NewExecutor(options ...cfg.Option[ExecutorConfig]) (Executor, error) {
	config := cfg.New(options...)

	exec, err := newExecutable(config)
	if err != nil {
		return noOpExecutable{}, err
	}

	if config.metrics != nil {
		exec = executorWithMetrics(exec, config.metrics)
	}

	if config.logger != nil {
		exec = executorWithLogs(exec, config.logger)
	}

	if config.tracer != nil {
		exec = executorWithTrace(exec, config.tracer)
	}

	return exec, nil
}

func newExecutable(config ExecutorConfig) (Executor, error) {
	// validate input
	if len(config.runners) == 0 {
		return noOpExecutable{}, ErrEmptyRunnerList
	}

	if config.scheduler == nil && config.cronString == "" {
		return noOpExecutable{}, ErrEmptyScheduler
	}

	var sched schedule.Scheduler

	// create a new schedule from a string if provided explicitly
	if config.cronString != "" {
		cron, err := schedule.New(schedule.WithSchedule(config.cronString))
		if err != nil {
			return noOpExecutable{}, err
		}

		sched = cron
	}

	// replace the scheduler with the provided one always
	if config.scheduler != nil {
		sched = config.scheduler
	}

	// prioritize the provided location by setting it last, even if the options
	// calls are not exactly optimized
	if config.loc != nil && sched != nil {
		cron, err := schedule.From(sched, schedule.WithLocation(config.loc))
		if err != nil {
			return noOpExecutable{}, err
		}

		sched = cron
	}

	// return the object with the provided runners
	return Executable{
		cron:    sched,
		runners: config.runners,
	}, nil
}

type noOpExecutable struct{}

func (e noOpExecutable) Exec(_ context.Context) error {
	return nil
}

func (e noOpExecutable) Next(_ context.Context) (t time.Time) {
	return t
}
