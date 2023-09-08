package schedule

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type SchedulerWithTrace struct {
	s      Scheduler
	tracer trace.Tracer
}

func (s SchedulerWithTrace) Next(ctx context.Context, now time.Time) time.Time {
	ctx, span := s.tracer.Start(ctx, "Scheduler.Next")
	defer span.End()

	next := s.s.Next(ctx, now)

	span.SetAttributes(attribute.String("at", next.Format(time.RFC3339)))

	return next
}

func withTrace(s Scheduler, tracer trace.Tracer) Scheduler {
	if s == nil {
		return noOpScheduler{}
	}

	if tracer == nil {
		return s
	}

	return SchedulerWithTrace{
		s:      s,
		tracer: tracer,
	}
}
