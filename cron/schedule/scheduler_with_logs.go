package schedule

import (
	"context"
	"log/slog"
	"os"
	"time"
)

type SchedulerWithLogs struct {
	s      Scheduler
	logger *slog.Logger
}

func (s SchedulerWithLogs) Next(ctx context.Context, now time.Time) time.Time {
	next := s.s.Next(ctx, now)

	s.logger.InfoContext(ctx, "next job", slog.Time("at", next))

	return next
}

func schedulerWithLogs(s Scheduler, handler slog.Handler) Scheduler {
	if s == nil {
		return noOpScheduler{}
	}

	if handler == nil {
		handler = slog.NewTextHandler(os.Stderr, nil)
	}

	if withLogs, ok := s.(SchedulerWithLogs); ok {
		withLogs.logger = slog.New(handler)

		return withLogs
	}

	return SchedulerWithLogs{
		s:      s,
		logger: slog.New(handler),
	}
}
