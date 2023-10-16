package executor

import (
	"context"
	"log/slog"
	"os"
	"time"
)

type ExecutorWithLogs struct {
	e      Executor
	logger *slog.Logger
}

func (e ExecutorWithLogs) Exec(ctx context.Context) error {
	id := slog.String("id", e.e.ID())

	e.logger.InfoContext(ctx, "executing task", id)

	err := e.e.Exec(ctx)
	if err != nil {
		e.logger.WarnContext(ctx, "task raised an error", id, slog.String("error", err.Error()))
	}

	return err
}

func (e ExecutorWithLogs) Next(ctx context.Context) time.Time {
	next := e.e.Next(ctx)

	e.logger.InfoContext(ctx, "next job",
		slog.String("id", e.e.ID()),
		slog.Time("at", next),
	)

	return next
}

func (e ExecutorWithLogs) ID() string {
	return e.e.ID()
}

func executorWithLogs(e Executor, handler slog.Handler) Executor {
	if e == nil {
		return noOpExecutor{}
	}

	if handler == nil {
		handler = slog.NewTextHandler(os.Stderr, nil)
	}

	if withLogs, ok := e.(ExecutorWithLogs); ok {
		withLogs.logger = slog.New(handler)

		return withLogs
	}

	return ExecutorWithLogs{
		e:      e,
		logger: slog.New(handler),
	}
}
