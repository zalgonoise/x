package reg

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestExample(t *testing.T) {
	// this is not an actual test, but a way of trying the API directly
	t.Skip()

	tracer := noop.NewTracerProvider().Tracer("test")

	r := New(nil, nil, nil)

	ctx := context.Background()
	ctx, span := tracer.Start(ctx, "test")
	defer span.End()

	metricFunc := func() {
		// increment something
	}

	err := errors.New("hello error")

	r.Event(ctx, "test event",
		WithError(err),
		WithLogAttributes(slog.String("user", "me")),
		WithSpan(span, attribute.String("user", "me")),
		WithMetric(metricFunc),
	)
}
