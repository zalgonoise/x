package service

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var _ Service = HandlerWithTrace{}

type HandlerWithTrace struct {
	s      Service
	tracer trace.Tracer
}

func (h HandlerWithTrace) Handle(ctx context.Context, value int) (err error) {
	ctx, span := h.tracer.Start(ctx, "handle",
		trace.WithAttributes(attribute.Int("value", value)),
	)

	defer span.End()

	if err = h.s.Handle(ctx, value); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return err
	}

	return nil
}

func WithTrace(s Service, tracer trace.Tracer) HandlerWithTrace {
	return HandlerWithTrace{
		s:      s,
		tracer: tracer,
	}
}
