package executor

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type withTrace struct {
	e      Executor
	tracer trace.Tracer
}

func (e withTrace) Exec(ctx context.Context) error {
	ctx, span := e.tracer.Start(ctx, "Executor.Exec")
	defer span.End()

	span.SetAttributes(attribute.String("id", e.e.ID()))

	err := e.e.Exec(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}

func (e withTrace) Next(ctx context.Context) time.Time {
	ctx, span := e.tracer.Start(ctx, "Executor.Next")
	defer span.End()

	next := e.e.Next(ctx)

	span.SetAttributes(
		attribute.String("id", e.e.ID()),
		attribute.String("at", next.Format(time.RFC3339)),
	)

	return next
}

func (e withTrace) ID() string {
	return e.e.ID()
}

func executorWithTrace(e Executor, tracer trace.Tracer) Executor {
	if e == nil {
		return noOpExecutor{}
	}

	if tracer == nil {
		return e
	}

	if traced, ok := e.(withTrace); ok {
		traced.tracer = tracer

		return traced
	}

	return withTrace{
		e:      e,
		tracer: tracer,
	}
}
