package selector

import (
	"context"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type SelectorWithTrace struct {
	s      Selector
	tracer trace.Tracer
}

func (s SelectorWithTrace) Next(ctx context.Context) error {
	ctx, span := s.tracer.Start(ctx, "Selector.Select")
	defer span.End()

	if err := s.s.Next(ctx); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return err
	}

	return nil
}

func selectorWithTrace(s Selector, tracer trace.Tracer) Selector {
	if s == nil {
		return noOpSelector{}
	}

	if tracer == nil {
		return s
	}

	return SelectorWithTrace{
		s:      s,
		tracer: tracer,
	}
}
