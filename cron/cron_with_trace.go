package cron

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type withTrace struct {
	r      Runtime
	tracer trace.Tracer
}

func (c withTrace) Run(ctx context.Context) {
	ctx, span := c.tracer.Start(ctx, "Runtime.Run")
	defer span.End()

	c.r.Run(ctx)

	span.AddEvent("closing runtime")
}

func (c withTrace) Err() <-chan error {
	return c.r.Err()
}

func cronWithTrace(r Runtime, tracer trace.Tracer) Runtime {
	if r == nil {
		return noOpRuntime{}
	}

	if tracer == nil {
		return r
	}

	if traced, ok := r.(withTrace); ok {
		traced.tracer = tracer

		return traced
	}

	return withTrace{
		r:      r,
		tracer: tracer,
	}
}
