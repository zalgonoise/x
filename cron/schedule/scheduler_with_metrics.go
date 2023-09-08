package schedule

import (
	"context"
	"time"
)

type Metrics interface {
	IncNextCalls()
}

type SchedulerWithMetrics struct {
	s Scheduler
	m Metrics
}

func (s SchedulerWithMetrics) Next(ctx context.Context, now time.Time) time.Time {
	s.m.IncNextCalls()

	return s.s.Next(ctx, now)
}

func withMetrics(s Scheduler, m Metrics) Scheduler {
	if s == nil {
		return noOpScheduler{}
	}

	if m == nil {
		return s
	}

	return SchedulerWithMetrics{
		s: s,
		m: m,
	}
}
