package schedule

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type withTrace struct {
	s      Scheduler
	tracer trace.Tracer
}

func (s withTrace) Next(ctx context.Context, now time.Time) time.Time {
	ctx, span := s.tracer.Start(ctx, "Scheduler.Next")
	defer span.End()

	next := s.s.Next(ctx, now)

	span.SetAttributes(attribute.String("at", next.Format(time.RFC3339)))

	return next
}

func schedulerWithTrace(s Scheduler, tracer trace.Tracer) Scheduler {
	if s == nil {
		return noOpScheduler{}
	}

	if tracer == nil {
		return s
	}

	if traced, ok := s.(withTrace); ok {
		traced.tracer = tracer

		return traced
	}

	return withTrace{
		s:      s,
		tracer: tracer,
	}
}
