package selector

import (
	"context"
	"time"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/errs"

	"github.com/zalgonoise/x/cron/executor"
)

const (
	minStepDuration = 50 * time.Millisecond

	errSelectorDomain = errs.Domain("x/cron/selector")

	ErrEmpty = errs.Kind("empty")

	ErrExecutorsList = errs.Entity("executors list")
)

var (
	ErrEmptyExecutorsList = errs.WithDomain(errSelectorDomain, ErrEmpty, ErrExecutorsList)
)

type Selector interface {
	Next(ctx context.Context) error
}

type selector struct {
	exec []executor.Executor
}

func (s selector) Next(ctx context.Context) error {
	// minStepDuration ensures that each execution is locked to the seconds mark and
	// a runner is not executed more than once per trigger.
	defer time.Sleep(minStepDuration)

	switch len(s.exec) {
	case 0:
		return ErrEmptyExecutorsList
	case 1:
		return s.exec[0].Exec(ctx)
	default:
		return executor.Multi(ctx, s.next(ctx)...)
	}
}

func (s selector) next(ctx context.Context) []executor.Executor {
	var (
		next time.Duration
		exec = make([]executor.Executor, 0, len(s.exec))
		now  = time.Now()
	)

	for i := range s.exec {
		t := s.exec[i].Next(ctx).Sub(now)

		switch {
		case i == 0:
			next = t
			exec = append(exec, s.exec[i])

			continue
		case t == next:
			exec = append(exec, s.exec[i])

			continue
		case t < next:
			next = t
			exec = make([]executor.Executor, 0, len(s.exec))
			exec = append(exec, s.exec[i])

			continue
		}
	}

	return exec
}

func New(options ...cfg.Option[Config]) (Selector, error) {
	config := cfg.New(options...)

	sel, err := newSelector(config)
	if err != nil {
		return noOpSelector{}, err
	}

	if config.metrics != nil {
		sel = selectorWithMetrics(sel, config.metrics)
	}

	if config.handler != nil {
		sel = selectorWithLogs(sel, config.handler)
	}

	if config.tracer != nil {
		sel = selectorWithTrace(sel, config.tracer)
	}

	return sel, nil
}

func newSelector(config Config) (Selector, error) {
	if len(config.exec) == 0 {
		return noOpSelector{}, ErrEmptyExecutorsList
	}

	return selector{
		exec: config.exec,
	}, nil
}

func NoOp() Selector {
	return noOpSelector{}
}

type noOpSelector struct{}

func (noOpSelector) Next(_ context.Context) error {
	return nil
}
