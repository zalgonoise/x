package logbuf

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/trace"
)

var ErrInvalidSpanContext = errors.New("input context does not contain a valid span context")

func GetTraceID(ctx context.Context) (trace.TraceID, error) {
	sc := trace.SpanContextFromContext(ctx)

	if sc.IsValid() {
		return sc.TraceID(), nil
	}

	return trace.TraceID{}, ErrInvalidSpanContext
}
