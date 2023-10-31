package schedule

import (
	"context"
	"time"
)

type Metrics interface {
	IncSchedulerNextCalls()
}

type withMetrics struct {
	s Scheduler
	m Metrics
}

func (s withMetrics) Next(ctx context.Context, now time.Time) time.Time {
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

	if metrics, ok := s.(withMetrics); ok {
		metrics.m = m

		return metrics
	}

	return withMetrics{
		s: s,
		m: m,
	}
}
