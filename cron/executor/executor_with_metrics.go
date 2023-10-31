package executor

import (
	"context"
	"time"
)

type Metrics interface {
	IncExecutorExecCalls(id string)
	IncExecutorExecErrors(id string)
	ObserveExecLatency(ctx context.Context, id string, dur time.Duration)
	IncExecutorNextCalls(id string)
}

type withMetrics struct {
	e Executor
	m Metrics
}

func (e withMetrics) Exec(ctx context.Context) error {
	id := e.e.ID()
	e.m.IncExecutorExecCalls(id)

	before := time.Now()

	err := e.e.Exec(ctx)

	e.m.ObserveExecLatency(ctx, id, time.Now().Sub(before))

	if err != nil {
		e.m.IncExecutorExecErrors(id)
	}

	return err
}

func (e withMetrics) Next(ctx context.Context) time.Time {
	e.m.IncExecutorNextCalls(e.e.ID())

	return e.e.Next(ctx)
}

func (e withMetrics) ID() string {
	return e.e.ID()
}

func executorWithMetrics(e Executor, m Metrics) Executor {
	if e == nil {
		return noOpExecutor{}
	}

	if m == nil {
		return e
	}

	if metrics, ok := e.(withMetrics); ok {
		metrics.m = m

		return metrics
	}

	return withMetrics{
		e: e,
		m: m,
	}
}
