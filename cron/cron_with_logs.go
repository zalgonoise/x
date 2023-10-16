package cron

import (
	"context"
	"log/slog"
	"os"
)

type CronWithLogs struct {
	r      Runtime
	logger *slog.Logger
}

func (c CronWithLogs) Run(ctx context.Context) {
	c.logger.InfoContext(ctx, "starting cron")
	c.r.Run(ctx)
	c.logger.InfoContext(ctx, "closing cron")
}

func (c CronWithLogs) Err() <-chan error {
	return c.r.Err()
}

func cronWithLogs(r Runtime, handler slog.Handler) Runtime {
	if r == nil {
		return noOpRuntime{}
	}

	if handler == nil {
		handler = slog.NewTextHandler(os.Stderr, nil)
	}

	if withLogs, ok := r.(CronWithLogs); ok {
		withLogs.logger = slog.New(handler)

		return withLogs
	}

	return CronWithLogs{
		r:      r,
		logger: slog.New(handler),
	}
}
