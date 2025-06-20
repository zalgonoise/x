package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

const (
	ServiceName = "collide"
)

// Tracer returns the registered tracer for this service. It defaults to a no-op trace.Tracer if not yet initialized.
func Tracer(name string) trace.Tracer {
	return otel.GetTracerProvider().Tracer(name)
}

type ShutdownFunc func(ctx context.Context) error

func Init(ctx context.Context, traceExporter sdktrace.SpanExporter) (ShutdownFunc, error) {
	if traceExporter == nil {
		otel.SetTracerProvider(noop.NewTracerProvider())

		return func(context.Context) error { return nil }, nil
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(ServiceName)), // the service name used to display traces in backends
	)
	if err != nil {
		return nil, err
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	// the tracer can now be referenced by the service name with a `otel.GetTracerProvider().Tracer(ServiceName)` call
	otel.SetTracerProvider(tracerProvider)

	// global propagator can't be the messaging propagator, otherwise gRPC propagation does not work
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Shutdown will flush any remaining spans and shut down the exporter.
	return tracerProvider.Shutdown, nil
}
