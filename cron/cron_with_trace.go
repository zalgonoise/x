package cron

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type CronWithTrace struct {
	r      Runtime
	tracer trace.Tracer
}

func (c CronWithTrace) Run(ctx context.Context) {
	ctx, span := c.tracer.Start(ctx, "Runtime.Run")
	defer span.End()

	c.r.Run(ctx)

	span.AddEvent("closing runtime")
}

func (c CronWithTrace) Err() <-chan error {
	return c.r.Err()
}

func cronWithTrace(r Runtime, tracer trace.Tracer) Runtime {
	if r == nil {
		return noOpRuntime{}
	}

	if tracer == nil {
		return r
	}

	if withTrace, ok := r.(CronWithTrace); ok {
		withTrace.tracer = tracer

		return withTrace
	}

	return CronWithTrace{
		r:      r,
		tracer: tracer,
	}
}
