package executor

import (
	"context"
	"log/slog"
	"os"
	"time"
)

type withLogs struct {
	e      Executor
	logger *slog.Logger
}

func (e withLogs) Exec(ctx context.Context) error {
	id := slog.String("id", e.e.ID())

	e.logger.InfoContext(ctx, "executing task", id)

	err := e.e.Exec(ctx)
	if err != nil {
		e.logger.WarnContext(ctx, "task raised an error", id, slog.String("error", err.Error()))
	}

	return err
}

func (e withLogs) Next(ctx context.Context) time.Time {
	next := e.e.Next(ctx)

	e.logger.InfoContext(ctx, "next job",
		slog.String("id", e.e.ID()),
		slog.Time("at", next),
	)

	return next
}

func (e withLogs) ID() string {
	return e.e.ID()
}

func executorWithLogs(e Executor, handler slog.Handler) Executor {
	if e == nil {
		return noOpExecutor{}
	}

	if handler == nil {
		handler = slog.NewTextHandler(os.Stderr, nil)
	}

	if logs, ok := e.(withLogs); ok {
		logs.logger = slog.New(handler)

		return logs
	}

	return withLogs{
		e:      e,
		logger: slog.New(handler),
	}
}
