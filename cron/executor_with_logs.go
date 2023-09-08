package cron

import (
	"context"
	"log/slog"
	"time"
)

type ExecutorWithLogs struct {
	e      Executor
	logger *slog.Logger
}

func (e ExecutorWithLogs) Exec(ctx context.Context) error {
	e.logger.InfoContext(ctx, "executing task")

	err := e.e.Exec(ctx)
	if err != nil {
		e.logger.WarnContext(ctx, "task raised an error", slog.String("error", err.Error()))
	}

	return err
}

func (e ExecutorWithLogs) Next(ctx context.Context) time.Time {
	next := e.e.Next(ctx)

	e.logger.InfoContext(ctx, "next job", slog.Time("at", next))

	return next
}

func executorWithLogs(e Executor, logger *slog.Logger) Executor {
	if e == nil {
		return noOpExecutable{}
	}

	if logger == nil {
		return e
	}

	return ExecutorWithLogs{
		e:      e,
		logger: logger,
	}
}
