package cron

import (
	"context"
	"log/slog"
	"os"
)

type withLogs struct {
	r      Runtime
	logger *slog.Logger
}

func (c withLogs) Run(ctx context.Context) {
	c.logger.InfoContext(ctx, "starting cron")
	c.r.Run(ctx)
	c.logger.InfoContext(ctx, "closing cron")
}

func (c withLogs) Err() <-chan error {
	return c.r.Err()
}

func cronWithLogs(r Runtime, handler slog.Handler) Runtime {
	if r == nil {
		return noOpRuntime{}
	}

	if handler == nil {
		handler = slog.NewTextHandler(os.Stderr, nil)
	}

	if logs, ok := r.(withLogs); ok {
		logs.logger = slog.New(handler)

		return logs
	}

	return withLogs{
		r:      r,
		logger: slog.New(handler),
	}
}
