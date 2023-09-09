package executor

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type ExecutorWithTrace struct {
	e      Executor
	tracer trace.Tracer
}

func (e ExecutorWithTrace) Exec(ctx context.Context) error {
	ctx, span := e.tracer.Start(ctx, "Executor.Exec")
	defer span.End()

	err := e.e.Exec(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}

func (e ExecutorWithTrace) Next(ctx context.Context) time.Time {
	ctx, span := e.tracer.Start(ctx, "Executor.Next")
	defer span.End()

	next := e.e.Next(ctx)

	span.SetAttributes(attribute.String("at", next.Format(time.RFC3339)))

	return next
}

func executorWithTrace(e Executor, tracer trace.Tracer) Executor {
	if e == nil {
		return noOpExecutable{}
	}

	if tracer == nil {
		return e
	}

	return ExecutorWithTrace{
		e:      e,
		tracer: tracer,
	}
}
