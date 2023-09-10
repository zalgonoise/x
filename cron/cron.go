package cron

import (
	"context"

	"github.com/zalgonoise/x/cfg"
	"github.com/zalgonoise/x/cron/executor"
	"github.com/zalgonoise/x/cron/selector"
)

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

				// filled error buffer; avoid dead-lock
				if len(r.err) == cap(r.err) {
					close(r.err)

					return
				}
			}
		}
	}
}

func (r runtime) Err() <-chan error {
	return r.err
}

func New(options ...cfg.Option[Config]) (Runtime, error) {
	config := cfg.New(options...)

	cron, err := newRuntime(config)
	if err != nil {
		return noOpRuntime{}, err
	}

	if config.metrics != nil {
		cron = cronWithMetrics(cron, config.metrics)
	}

	if config.logger != nil {
		cron = cronWithLogs(cron, config.logger)
	}

	if config.tracer != nil {
		cron = cronWithTrace(cron, config.tracer)
	}

	return cron, nil
}

func newRuntime(config Config) (Runtime, error) {
	// validate input
	if config.sel == nil {
		return noOpRuntime{}, executor.ErrEmptySelector
	}

	size := config.errBufferSize
	if size < minBufferSize {
		size = defaultBufferSize
	}

	return runtime{
		sel: config.sel,
		err: make(chan error, size),
	}, nil
}

type noOpRuntime struct{}

func (noOpRuntime) Run(context.Context) {}

func (noOpRuntime) Err() <-chan error {
	return nil
}
