package executor

import (
	"context"
	"time"
)

type ExecutorMetrics interface {
	IncExecutorExecCalls(id string)
	IncExecutorExecErrors(id string)
	ObserveExecLatency(ctx context.Context, id string, dur time.Duration)
	IncExecutorNextCalls(id string)
}

type ExecutorWithMetrics struct {
	e Executor
	m ExecutorMetrics
}

func (e ExecutorWithMetrics) Exec(ctx context.Context) error {
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

func (e ExecutorWithMetrics) Next(ctx context.Context) time.Time {
	e.m.IncExecutorNextCalls(e.e.ID())

	return e.e.Next(ctx)
}

func (e ExecutorWithMetrics) ID() string {
	return e.e.ID()
}

func executorWithMetrics(e Executor, m ExecutorMetrics) Executor {
	if e == nil {
		return noOpExecutable{}
	}

	if m == nil {
		return e
	}

	return ExecutorWithMetrics{
		e: e,
		m: m,
	}
}
