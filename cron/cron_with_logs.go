package cron

import (
	"context"
	"log/slog"
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

func cronWithLogs(r Runtime, logger *slog.Logger) Runtime {
	if r == nil {
		return noOpRuntime{}
	}

	if logger == nil {
		return r
	}

	return CronWithLogs{
		r:      r,
		logger: logger,
	}
}
