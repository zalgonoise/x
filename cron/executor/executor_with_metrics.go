package executor

import (
	"context"
	"time"
)

type ExecutorMetrics interface {
	IncExecCalls()
	IncExecErrors()
	IncNextCalls()
}

type ExecutorWithMetrics struct {
	e Executor
	m ExecutorMetrics
}

func (e ExecutorWithMetrics) Exec(ctx context.Context) error {
	e.m.IncNextCalls()

	err := e.e.Exec(ctx)
	if err != nil {
		e.m.IncExecErrors()
	}

	return err
}

func (e ExecutorWithMetrics) Next(ctx context.Context) time.Time {
	e.m.IncNextCalls()

	return e.e.Next(ctx)
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
