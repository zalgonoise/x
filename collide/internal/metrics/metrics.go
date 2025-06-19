package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"go.opentelemetry.io/otel/trace"
)

const (
	// traceIDKey is used as the trace ID key value in the prometheus.Labels in a prometheus.Exemplar.
	//
	// Its value of `trace_id` complies with the OpenTelemetry specification for metrics' exemplars, as seen in:
	// https://opentelemetry.io/docs/specs/otel/metrics/data-model/#exemplars
	traceIDKey = "trace_id"
)

type Metrics struct {
	// CollideService metrics
	listDistrictsTotal          prometheus.Counter
	listDistrictsFailed         prometheus.Counter
	listDistrictsLatencySeconds prometheus.Histogram

	listAllTracksByDistrictTotal          *prometheus.CounterVec
	listAllTracksByDistrictFailed         *prometheus.CounterVec
	listAllTracksByDistrictLatencySeconds *prometheus.HistogramVec

	listDriftTracksByDistrictTotal          *prometheus.CounterVec
	listDriftTracksByDistrictFailed         *prometheus.CounterVec
	listDriftTracksByDistrictLatencySeconds *prometheus.HistogramVec

	getAlternativesByDistrictAndTrackTotal          *prometheus.CounterVec
	getAlternativesByDistrictAndTrackFailed         *prometheus.CounterVec
	getAlternativesByDistrictAndTrackLatencySeconds *prometheus.HistogramVec

	getCollisionsByDistrictAndTrackTotal          *prometheus.CounterVec
	getCollisionsByDistrictAndTrackFailed         *prometheus.CounterVec
	getCollisionsByDistrictAndTrackLatencySeconds *prometheus.HistogramVec

	// Third party metrics
	collectors []prometheus.Collector
}

func NewMetrics() *Metrics {
	return &Metrics{
		listDistrictsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "list_districts_total",
			Help: "Count of requests to list districts",
		}),
		listDistrictsFailed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "list_districts_failed",
			Help: "Count of failed requests to list districts",
		}),
		listDistrictsLatencySeconds: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "list_districts_latency_seconds",
			Help:    "Latency of requests to list districts",
			Buckets: []float64{.00001, .00005, .0001, .0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}),

		listAllTracksByDistrictTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "list_all_tracks_by_district_total",
			Help: "Count of requests to list tracks within a certain district",
		}, []string{"district"}),
		listAllTracksByDistrictFailed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "list_all_tracks_by_district_failed",
			Help: "Count of failed requests to list tracks within a certain district",
		}, []string{"district"}),
		listAllTracksByDistrictLatencySeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "list_all_tracks_by_district_latency_seconds",
			Help:    "Latency of requests to list all tracks within a certain district",
			Buckets: []float64{.00001, .00005, .0001, .0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}, []string{"district"}),

		listDriftTracksByDistrictTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "list_drift_tracks_by_district_total",
			Help: "Count of requests to list drift tracks within a certain district",
		}, []string{"district"}),
		listDriftTracksByDistrictFailed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "list_drift_tracks_by_district_failed",
			Help: "Count of failed requests to list drift tracks within a certain district",
		}, []string{"district"}),
		listDriftTracksByDistrictLatencySeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "list_drift_tracks_by_district_latency_seconds",
			Help:    "Latency of requests to list drift tracks within a certain district",
			Buckets: []float64{.00001, .00005, .0001, .0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}, []string{"district"}),

		getAlternativesByDistrictAndTrackTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "get_alternatives_by_district_and_track_total",
			Help: "Count of requests to get alternatives for a certain district, with a target track",
		}, []string{"district", "track"}),
		getAlternativesByDistrictAndTrackFailed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "get_alternatives_by_district_and_track_failed",
			Help: "Count of failed requests to get alternatives for a certain district, with a target track",
		}, []string{"district", "track"}),
		getAlternativesByDistrictAndTrackLatencySeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "get_alternatives_by_district_and_track_latency_seconds",
			Help:    "Latency of request to get alternatives for a certain district, with a target track",
			Buckets: []float64{.00001, .00005, .0001, .0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}, []string{"district", "track"}),

		getCollisionsByDistrictAndTrackTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "get_collisions_by_district_and_track_total",
			Help: "Count of requests to get collisions for a certain district, with a target track",
		}, []string{"district", "track"}),
		getCollisionsByDistrictAndTrackFailed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "get_collisions_by_district_and_track_failed",
			Help: "Count of failed requests to get collisions for a certain district, with a target track",
		}, []string{"district", "track"}),
		getCollisionsByDistrictAndTrackLatencySeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "get_collisions_by_district_and_track_latency_seconds",
			Help:    "Latency of request to get collisions for a certain district, with a target track",
			Buckets: []float64{.00001, .00005, .0001, .0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}, []string{"district", "track"}),
	}
}

func (m *Metrics) IncListDistricts()       { m.listDistrictsTotal.Inc() }
func (m *Metrics) IncListDistrictsFailed() { m.listDistrictsFailed.Inc() }
func (m *Metrics) ObserveListDistrictsLatency(ctx context.Context, duration time.Duration) {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		if eo, ok := m.listDistrictsLatencySeconds.(prometheus.ExemplarObserver); ok {
			eo.ObserveWithExemplar(duration.Seconds(), prometheus.Labels{
				traceIDKey: sc.TraceID().String(),
			})

			return
		}
	}

	m.listDistrictsLatencySeconds.Observe(duration.Seconds())
}
func (m *Metrics) IncListAllTracksByDistrict(district string) {
	m.listAllTracksByDistrictTotal.WithLabelValues(district).Inc()
}
func (m *Metrics) IncListAllTracksByDistrictFailed(district string) {
	m.listAllTracksByDistrictFailed.WithLabelValues(district).Inc()
}
func (m *Metrics) ObserveListAllTracksByDistrictLatency(ctx context.Context, duration time.Duration, district string) {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		if eo, ok := m.listAllTracksByDistrictLatencySeconds.
			WithLabelValues(district).(prometheus.ExemplarObserver); ok {
			eo.ObserveWithExemplar(duration.Seconds(), prometheus.Labels{
				traceIDKey: sc.TraceID().String(),
			})

			return
		}
	}

	m.listAllTracksByDistrictLatencySeconds.WithLabelValues(district).Observe(duration.Seconds())
}
func (m *Metrics) IncListDriftTracksByDistrict(district string) {
	m.listDriftTracksByDistrictTotal.WithLabelValues(district).Inc()
}
func (m *Metrics) IncListDriftTracksByDistrictFailed(district string) {
	m.listDriftTracksByDistrictFailed.WithLabelValues(district).Inc()
}
func (m *Metrics) ObserveListDriftTracksByDistrictLatency(ctx context.Context, duration time.Duration, district string) {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		if eo, ok := m.listDriftTracksByDistrictLatencySeconds.
			WithLabelValues(district).(prometheus.ExemplarObserver); ok {
			eo.ObserveWithExemplar(duration.Seconds(), prometheus.Labels{
				traceIDKey: sc.TraceID().String(),
			})

			return
		}
	}

	m.listDriftTracksByDistrictLatencySeconds.WithLabelValues(district).Observe(duration.Seconds())
}
func (m *Metrics) IncGetAlternativesByDistrictAndTrack(district, track string) {
	m.getAlternativesByDistrictAndTrackTotal.WithLabelValues(district, track).Inc()
}
func (m *Metrics) IncGetAlternativesByDistrictAndTrackFailed(district, track string) {
	m.getAlternativesByDistrictAndTrackFailed.WithLabelValues(district, track).Inc()
}
func (m *Metrics) ObserveGetAlternativesByDistrictAndTrackLatency(ctx context.Context, duration time.Duration, district, track string) {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		if eo, ok := m.getAlternativesByDistrictAndTrackLatencySeconds.
			WithLabelValues(district, track).(prometheus.ExemplarObserver); ok {
			eo.ObserveWithExemplar(duration.Seconds(), prometheus.Labels{
				traceIDKey: sc.TraceID().String(),
			})

			return
		}
	}

	m.getAlternativesByDistrictAndTrackLatencySeconds.WithLabelValues(district, track).Observe(duration.Seconds())
}
func (m *Metrics) IncGetCollisionsByDistrictAndTrack(district, track string) {
	m.getCollisionsByDistrictAndTrackTotal.WithLabelValues(district, track).Inc()
}
func (m *Metrics) IncGetCollisionsByDistrictAndTrackFailed(district, track string) {
	m.getCollisionsByDistrictAndTrackFailed.WithLabelValues(district, track).Inc()
}
func (m *Metrics) ObserveGetCollisionsByDistrictAndTrackLatency(ctx context.Context, duration time.Duration, district, track string) {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		if eo, ok := m.getCollisionsByDistrictAndTrackLatencySeconds.
			WithLabelValues(district, track).(prometheus.ExemplarObserver); ok {
			eo.ObserveWithExemplar(duration.Seconds(), prometheus.Labels{
				traceIDKey: sc.TraceID().String(),
			})

			return
		}
	}

	m.getCollisionsByDistrictAndTrackLatencySeconds.WithLabelValues(district, track).Observe(duration.Seconds())
}

func (m *Metrics) RegisterCollector(collector prometheus.Collector) {
	m.collectors = append(m.collectors, collector)
}

func (m *Metrics) Registry() (*prometheus.Registry, error) {
	reg := prometheus.NewRegistry()

	for _, metric := range []prometheus.Collector{
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{
			ReportErrors: false,
		}),
		m.listDistrictsTotal,
		m.listDistrictsFailed,
		m.listDistrictsLatencySeconds,

		m.listAllTracksByDistrictTotal,
		m.listAllTracksByDistrictFailed,
		m.listAllTracksByDistrictLatencySeconds,

		m.listDriftTracksByDistrictTotal,
		m.listDriftTracksByDistrictFailed,
		m.listDriftTracksByDistrictLatencySeconds,

		m.getAlternativesByDistrictAndTrackTotal,
		m.getAlternativesByDistrictAndTrackFailed,
		m.getAlternativesByDistrictAndTrackLatencySeconds,

		m.getCollisionsByDistrictAndTrackTotal,
		m.getCollisionsByDistrictAndTrackFailed,
		m.getCollisionsByDistrictAndTrackLatencySeconds,
	} {
		err := reg.Register(metric)
		if err != nil {
			return nil, err
		}
	}

	for _, metric := range m.collectors {
		err := reg.Register(metric)
		if err != nil {
			return nil, err
		}
	}

	return reg, nil
}
