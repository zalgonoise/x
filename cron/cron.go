package cron

import (
	"context"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/cron/selector"
	"github.com/zalgonoise/x/errs"
)

const (
	errDomain = errs.Domain("x/cron")

	ErrEmpty = errs.Kind("empty")

	ErrSelector = errs.Entity("task selector")
)

var ErrEmptySelector = errs.WithDomain(errDomain, ErrEmpty, ErrSelector)

type Runtime interface {
	Run(ctx context.Context)
	Err() <-chan error
}

type runtime struct {
	sel selector.Selector

	err chan error
}

func (r runtime) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := r.sel.Next(ctx); err != nil {
				r.err <- err
			}
		}
	}
}

func (r runtime) Err() <-chan error {
	return r.err
}

func New(sel selector.Selector, options ...cfg.Option[Config]) (Runtime, error) {
	config := cfg.New(options...)

	cron, err := newRuntime(sel, config)
	if err != nil {
		return noOpRuntime{}, err
	}

	if config.metrics != nil {
		cron = cronWithMetrics(cron, config.metrics)
	}

	if config.handler != nil {
		cron = cronWithLogs(cron, config.handler)
	}

	if config.tracer != nil {
		cron = cronWithTrace(cron, config.tracer)
	}

	return cron, nil
}

func newRuntime(sel selector.Selector, config Config) (Runtime, error) {
	// validate input
	if sel == nil {
		return noOpRuntime{}, ErrEmptySelector
	}

	size := config.errBufferSize
	if size < minBufferSize {
		size = defaultBufferSize
	}

	return runtime{
		sel: sel,
		err: make(chan error, size),
	}, nil
}

func NoOp() Runtime {
	return noOpRuntime{}
}

type noOpRuntime struct{}

func (noOpRuntime) Run(context.Context) {}

func (noOpRuntime) Err() <-chan error {
	return nil
}
