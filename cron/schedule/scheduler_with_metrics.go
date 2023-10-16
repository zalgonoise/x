package schedule

import (
	"context"
	"time"
)

type Metrics interface {
	IncSchedulerNextCalls()
}

type SchedulerWithMetrics struct {
	s Scheduler
	m Metrics
}

func (s SchedulerWithMetrics) Next(ctx context.Context, now time.Time) time.Time {
	s.m.IncSchedulerNextCalls()

	return s.s.Next(ctx, now)
}

func schedulerWithMetrics(s Scheduler, m Metrics) Scheduler {
	if s == nil {
		return noOpScheduler{}
	}

	if m == nil {
		return s
	}

	if withMetrics, ok := s.(SchedulerWithMetrics); ok {
		withMetrics.m = m

		return withMetrics
	}

	return SchedulerWithMetrics{
		s: s,
		m: m,
	}
}
