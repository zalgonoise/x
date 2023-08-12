package metrics

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
)

const defaultInterval = 500 * time.Millisecond
const ServiceName = "exemplars"

type MetricsV2 struct {
	serviceUp metric.Int64UpDownCounter

	reqReceived metric.Int64Counter
	reqFailed   metric.Int64Counter
	reqLatency  metric.Float64Histogram
}

func (m *MetricsV2) IncRequestsReceived(ctx context.Context) {
	m.reqReceived.Add(ctx, 1)
}

func (m *MetricsV2) IncRequestsFailed(ctx context.Context) {
	m.reqFailed.Add(ctx, 1)
}

func (m *MetricsV2) ObserveHandlingLatency(ctx context.Context, dur time.Duration) {
	// OpenTelemetry currently does not support exemplars in Histograms.
	// This makes it so that Prometheus is key in this type of implementation
	//
	//if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
	//	m.reqLatency.Record(ctx, dur.Seconds(), metric.WithAttributeSet(
	//		attribute.NewSet(
	//			attribute.String(traceIDKey, sc.TraceID().String()),
	//		),
	//	))
	//
	//	return
	//}

	m.reqLatency.Record(ctx, dur.Seconds())
}

func NewMetricsV2() (*MetricsV2, error) {
	up, err := Meter().Int64UpDownCounter(
		"up",
		metric.WithDescription("indicates whether the service is running or not"),
	)
	if err != nil {
		return nil, err
	}

	recv, err := Meter().Int64Counter(
		"req_received_total",
		metric.WithUnit("req"),
		metric.WithDescription("Count of the requests received by the service handler"),
	)
	if err != nil {
		return nil, err
	}

	failed, err := Meter().Int64Counter(
		"req_failed_total",
		metric.WithUnit("req"),
		metric.WithDescription("Count of the requests that have failed, in the service handler"),
	)
	if err != nil {
		return nil, err
	}

	latency, err := Meter().Float64Histogram(
		"req_handling_latency_seconds",
		metric.WithUnit("s"),
		metric.WithDescription("Histogram of request handling latencies"),
	)
	if err != nil {
		return nil, err
	}

	return &MetricsV2{
		serviceUp:   up,
		reqReceived: recv,
		reqFailed:   failed,
		reqLatency:  latency,
	}, nil
}

type ShutdownFunc func(ctx context.Context) error

func Meter() metric.Meter {
	return otel.GetMeterProvider().Meter(ServiceName)
}

func Init(ctx context.Context, addr string) (ShutdownFunc, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(ServiceName)),
	)
	if err != nil {
		return nil, err
	}

	exporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint(addr),
		otlpmetrichttp.WithInsecure(),
		otlpmetrichttp.WithHeaders(map[string]string{
			"X-Scope-OrgID": "anonymous",
		}),
		otlpmetrichttp.WithRetry(otlpmetrichttp.RetryConfig{
			Enabled:         true,
			InitialInterval: 100 * time.Millisecond,
			MaxInterval:     500 * time.Millisecond,
			MaxElapsedTime:  time.Minute,
		}),
	)
	if err != nil {
		return nil, err
	}

	meterProvider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(
		exporter,
		sdkmetric.WithInterval(defaultInterval),
	)),
		sdkmetric.WithResource(res),
	)

	otel.SetMeterProvider(meterProvider)

	return meterProvider.Shutdown, nil
}
