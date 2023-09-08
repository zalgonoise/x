package schedule

import (
	"context"
	"log/slog"
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

func withLogs(s Scheduler, logger *slog.Logger) Scheduler {
	if s == nil {
		return noOpScheduler{}
	}

	if logger == nil {
		return s
	}

	return SchedulerWithLogs{
		s:      s,
		logger: logger,
	}
}
