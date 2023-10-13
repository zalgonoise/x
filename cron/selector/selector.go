package selector

import (
	"context"
	"time"

	"github.com/zalgonoise/x/cfg"
	"github.com/zalgonoise/x/cron/executor"
	"github.com/zalgonoise/x/errs"
)

const (
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
	var exec executor.Executor

	switch len(s.exec) {
	case 0:
		return ErrEmptyExecutorsList
	case 1:
		exec = s.exec[0]
	default:
		exec = s.next(ctx)
	}

	if err := exec.Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (s selector) next(ctx context.Context) executor.Executor {
	var (
		nextTime time.Time
		idx      int
	)

	for i := range s.exec {
		t := s.exec[i].Next(ctx)

		if i == 0 {
			nextTime = t

			continue
		}

		if t.Before(nextTime) {
			nextTime = t
			idx = i
		}
	}

	return s.exec[idx]
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

	if config.logger != nil {
		sel = selectorWithLogs(sel, config.logger)
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

type noOpSelector struct{}

func (noOpSelector) Next(_ context.Context) error {
	return nil
}
