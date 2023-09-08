package cron

import (
	"context"
	"time"

	"github.com/zalgonoise/x/cfg"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Runtime interface {
	Run(ctx context.Context)
	Err() <-chan error
}

type runtime struct {
	exec []Executor

	tracer trace.Tracer
	err    chan error
}

func (r runtime) Run(ctx context.Context) {
	var (
		runnerCtx context.Context
		cancel    context.CancelFunc
	)

	for {
		select {
		case <-ctx.Done():
			if cancel != nil {
				cancel()
			}

			return
		default:
			if cancel == nil {
				// TODO: refactor to include a collector (or w/e) type to wrap this section
				ctx, span := r.tracer.Start(ctx, "Runtime.Run")

				runnerCtx, cancel = context.WithCancel(ctx)

				idx := r.nextTaskIndex(runnerCtx)
				if idx < 0 {
					cancel()
					cancel = nil

					span.End()

					return
				}

				go func() {
					if err := r.exec[idx].Exec(runnerCtx); err != nil {
						span.SetStatus(codes.Error, err.Error())
						span.RecordError(err)

						r.err <- err
					}

					span.End()

					cancel()
					cancel = nil
				}()
			}
		}
	}
}

func (r runtime) Err() <-chan error {
	return r.err
}

func (r runtime) nextTaskIndex(ctx context.Context) int {
	if len(r.exec) < 1 {
		return -1
	}

	var (
		nextTime time.Time
		idx      int
	)

	for i := range r.exec {
		t := r.exec[i].Next(ctx)

		if i == 0 {
			nextTime = t

			continue
		}

		if t.Before(nextTime) {
			nextTime = t
			idx = i
		}
	}

	return idx
}

func New(options ...cfg.Option[RuntimeConfig]) (Runtime, error) {
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

	return cron, nil
}

func newRuntime(config RuntimeConfig) (Runtime, error) {
	// validate input
	if len(config.exec) == 0 {
		return noOpRuntime{}, ErrEmptyExecutableList
	}

	tracer := trace.NewNoopTracerProvider().Tracer("cron")
	if config.tracer != nil {
		tracer = config.tracer
	}

	return runtime{
		exec:   config.exec,
		tracer: tracer,
	}, nil
}

type noOpRuntime struct{}

func (noOpRuntime) Run(context.Context) {}

func (noOpRuntime) Err() <-chan error {
	return nil
}
