package metrics

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
)

const defaultInterval = 500 * time.Millisecond
const ServiceName = "collide"

type ShutdownFunc func(ctx context.Context) error

func Meter() metric.Meter {
	return otel.GetMeterProvider().Meter(ServiceName)
}

var bucketBoundaries = []float64{
	.00001, .00005, .0001, .0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10,
}

type Otel struct {
	// CollideService metrics
	listDistrictsTotal          metric.Int64Counter
	listDistrictsFailed         metric.Int64Counter
	listDistrictsLatencySeconds metric.Float64Histogram

	listAllTracksByDistrictTotal          metric.Int64Counter
	listAllTracksByDistrictFailed         metric.Int64Counter
	listAllTracksByDistrictLatencySeconds metric.Float64Histogram

	listDriftTracksByDistrictTotal          metric.Int64Counter
	listDriftTracksByDistrictFailed         metric.Int64Counter
	listDriftTracksByDistrictLatencySeconds metric.Float64Histogram

	getAlternativesByDistrictAndTrackTotal          metric.Int64Counter
	getAlternativesByDistrictAndTrackFailed         metric.Int64Counter
	getAlternativesByDistrictAndTrackLatencySeconds metric.Float64Histogram

	getCollisionsByDistrictAndTrackTotal          metric.Int64Counter
	getCollisionsByDistrictAndTrackFailed         metric.Int64Counter
	getCollisionsByDistrictAndTrackLatencySeconds metric.Float64Histogram
}

func NewOtel() (*Otel, error) {
	listDistrictsTotal, err := Meter().Int64Counter(
		"list_districts_total",
		metric.WithUnit("req"),
		metric.WithDescription("Count of requests to list districts"),
	)
	if err != nil {
		return nil, err
	}

	listDistrictsFailed, err := Meter().Int64Counter(
		"list_districts_failed",
		metric.WithUnit("req"),
		metric.WithDescription("Count of failed requests to list districts"),
	)
	if err != nil {
		return nil, err
	}

	listDistrictsLatencySeconds, err := Meter().Float64Histogram(
		"list_districts_latency_seconds",
		metric.WithUnit("s"),
		metric.WithDescription("Latency of requests to list districts"),
		metric.WithExplicitBucketBoundaries(bucketBoundaries...))
	if err != nil {
		return nil, err
	}

	listAllTracksByDistrictTotal, err := Meter().Int64Counter(
		"list_all_tracks_by_district_total",
		metric.WithUnit("req"),
		metric.WithDescription("Count of requests to list tracks within a certain district"),
	)
	if err != nil {
		return nil, err
	}

	listAllTracksByDistrictFailed, err := Meter().Int64Counter(
		"list_all_tracks_by_district_failed",
		metric.WithUnit("req"),
		metric.WithDescription("Count of failed requests to list tracks within a certain district"),
	)
	if err != nil {
		return nil, err
	}

	listAllTracksByDistrictLatencySeconds, err := Meter().Float64Histogram(
		"list_all_tracks_by_district_latency_seconds",
		metric.WithUnit("s"),
		metric.WithDescription("Latency of requests to list all tracks within a certain district"),
		metric.WithExplicitBucketBoundaries(bucketBoundaries...))
	if err != nil {
		return nil, err
	}

	listDriftTracksByDistrictTotal, err := Meter().Int64Counter(
		"list_drift_tracks_by_district_total",
		metric.WithUnit("req"),
		metric.WithDescription("Count of requests to list drift tracks within a certain district"),
	)
	if err != nil {
		return nil, err
	}

	listDriftTracksByDistrictFailed, err := Meter().Int64Counter(
		"list_drift_tracks_by_district_failed",
		metric.WithUnit("req"),
		metric.WithDescription("Count of failed requests to list drift tracks within a certain district"),
	)
	if err != nil {
		return nil, err
	}

	listDriftTracksByDistrictLatencySeconds, err := Meter().Float64Histogram(
		"list_drift_tracks_by_district_latency_seconds",
		metric.WithUnit("s"),
		metric.WithDescription("Latency of requests to list drift tracks within a certain district"),
		metric.WithExplicitBucketBoundaries(bucketBoundaries...))
	if err != nil {
		return nil, err
	}

	getAlternativesByDistrictAndTrackTotal, err := Meter().Int64Counter(
		"get_alternatives_by_district_and_track_total",
		metric.WithUnit("req"),
		metric.WithDescription("Count of requests to get alternatives for a certain district, with a target track"),
	)
	if err != nil {
		return nil, err
	}

	getAlternativesByDistrictAndTrackFailed, err := Meter().Int64Counter(
		"get_alternatives_by_district_and_track_failed",
		metric.WithUnit("req"),
		metric.WithDescription("Count of failed requests to get alternatives for a certain district, with a target track"),
	)
	if err != nil {
		return nil, err
	}

	getAlternativesByDistrictAndTrackLatencySeconds, err := Meter().Float64Histogram(
		"get_alternatives_by_district_and_track_latency_seconds",
		metric.WithUnit("s"),
		metric.WithDescription("Latency of request to get alternatives for a certain district, with a target track"),
		metric.WithExplicitBucketBoundaries(bucketBoundaries...))
	if err != nil {
		return nil, err
	}

	getCollisionsByDistrictAndTrackTotal, err := Meter().Int64Counter(
		"get_collisions_by_district_and_track_total",
		metric.WithUnit("req"),
		metric.WithDescription("Count of requests to get collisions for a certain district, with a target track"),
	)
	if err != nil {
		return nil, err
	}

	getCollisionsByDistrictAndTrackFailed, err := Meter().Int64Counter(
		"get_collisions_by_district_and_track_failed",
		metric.WithUnit("req"),
		metric.WithDescription("Count of failed requests to get collisions for a certain district, with a target track"),
	)
	if err != nil {
		return nil, err
	}

	getCollisionsByDistrictAndTrackLatencySeconds, err := Meter().Float64Histogram(
		"get_collisions_by_district_and_track_latency_seconds",
		metric.WithUnit("s"),
		metric.WithDescription("Latency of request to get collisions for a certain district, with a target track"),
		metric.WithExplicitBucketBoundaries(bucketBoundaries...))
	if err != nil {
		return nil, err
	}

	return &Otel{
		listDistrictsTotal:          listDistrictsTotal,
		listDistrictsFailed:         listDistrictsFailed,
		listDistrictsLatencySeconds: listDistrictsLatencySeconds,

		listAllTracksByDistrictTotal:          listAllTracksByDistrictTotal,
		listAllTracksByDistrictFailed:         listAllTracksByDistrictFailed,
		listAllTracksByDistrictLatencySeconds: listAllTracksByDistrictLatencySeconds,

		listDriftTracksByDistrictTotal:          listDriftTracksByDistrictTotal,
		listDriftTracksByDistrictFailed:         listDriftTracksByDistrictFailed,
		listDriftTracksByDistrictLatencySeconds: listDriftTracksByDistrictLatencySeconds,

		getAlternativesByDistrictAndTrackTotal:          getAlternativesByDistrictAndTrackTotal,
		getAlternativesByDistrictAndTrackFailed:         getAlternativesByDistrictAndTrackFailed,
		getAlternativesByDistrictAndTrackLatencySeconds: getAlternativesByDistrictAndTrackLatencySeconds,

		getCollisionsByDistrictAndTrackTotal:          getCollisionsByDistrictAndTrackTotal,
		getCollisionsByDistrictAndTrackFailed:         getCollisionsByDistrictAndTrackFailed,
		getCollisionsByDistrictAndTrackLatencySeconds: getCollisionsByDistrictAndTrackLatencySeconds,
	}, nil
}

func (m *Otel) IncListDistricts(ctx context.Context) {
	m.listDistrictsTotal.Add(ctx, 1)
}
func (m *Otel) IncListDistrictsFailed(ctx context.Context) {
	m.listDistrictsFailed.Add(ctx, 1)
}
func (m *Otel) ObserveListDistrictsLatency(ctx context.Context, duration time.Duration) {
	m.listDistrictsLatencySeconds.Record(ctx, duration.Seconds())
}
func (m *Otel) IncListAllTracksByDistrict(ctx context.Context, district string) {
	m.listAllTracksByDistrictTotal.Add(ctx, 1, metric.WithAttributes(attribute.String("district", district)))
}
func (m *Otel) IncListAllTracksByDistrictFailed(ctx context.Context, district string) {
	m.listAllTracksByDistrictFailed.Add(ctx, 1, metric.WithAttributes(attribute.String("district", district)))
}
func (m *Otel) ObserveListAllTracksByDistrictLatency(ctx context.Context, duration time.Duration, district string) {
	m.listAllTracksByDistrictLatencySeconds.Record(ctx, duration.Seconds(), metric.WithAttributes(
		attribute.String("district", district)))
}
func (m *Otel) IncListDriftTracksByDistrict(ctx context.Context, district string) {
	m.listDriftTracksByDistrictTotal.Add(ctx, 1, metric.WithAttributes(attribute.String("district", district)))
}

func (m *Otel) IncListDriftTracksByDistrictFailed(ctx context.Context, district string) {
	m.listDriftTracksByDistrictFailed.Add(ctx, 1, metric.WithAttributes(attribute.String("district", district)))
}

func (m *Otel) ObserveListDriftTracksByDistrictLatency(ctx context.Context, duration time.Duration, district string) {
	m.listDriftTracksByDistrictLatencySeconds.Record(ctx, duration.Seconds(), metric.WithAttributes(
		attribute.String("district", district)))
}
func (m *Otel) IncGetAlternativesByDistrictAndTrack(ctx context.Context, district, track string) {
	m.getAlternativesByDistrictAndTrackTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("district", district),
		attribute.String("track", track),
	))
}
func (m *Otel) IncGetAlternativesByDistrictAndTrackFailed(ctx context.Context, district, track string) {
	m.getAlternativesByDistrictAndTrackFailed.Add(ctx, 1, metric.WithAttributes(
		attribute.String("district", district),
		attribute.String("track", track),
	))
}
func (m *Otel) ObserveGetAlternativesByDistrictAndTrackLatency(ctx context.Context, duration time.Duration, district, track string) {
	m.getAlternativesByDistrictAndTrackLatencySeconds.Record(ctx, duration.Seconds(), metric.WithAttributes(
		attribute.String("district", district),
		attribute.String("track", track),
	))
}
func (m *Otel) IncGetCollisionsByDistrictAndTrack(ctx context.Context, district, track string) {
	m.getCollisionsByDistrictAndTrackTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("district", district),
		attribute.String("track", track),
	))
}
func (m *Otel) IncGetCollisionsByDistrictAndTrackFailed(ctx context.Context, district, track string) {
	m.getCollisionsByDistrictAndTrackFailed.Add(ctx, 1, metric.WithAttributes(
		attribute.String("district", district),
		attribute.String("track", track),
	))
}
func (m *Otel) ObserveGetCollisionsByDistrictAndTrackLatency(ctx context.Context, duration time.Duration, district, track string) {
	m.getCollisionsByDistrictAndTrackLatencySeconds.Record(ctx, duration.Seconds(), metric.WithAttributes(
		attribute.String("district", district),
		attribute.String("track", track),
	))
}

func Init(ctx context.Context, uri string) (ShutdownFunc, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(ServiceName)),
	)
	if err != nil {
		return nil, err
	}

	exporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint(uri),
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
