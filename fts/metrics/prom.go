package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"go.opentelemetry.io/otel/trace"
)

func (m *Metrics) IncSearchesTotal() {
	m.searchesTotal.Inc()
}

func (m *Metrics) IncSearchesFailed() {
	m.searchesFailed.Inc()
}

func (m *Metrics) ObserveSearchLatency(ctx context.Context, dur time.Duration) {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		m.searchesLatency.(prometheus.ExemplarObserver).ObserveWithExemplar(dur.Seconds(), prometheus.Labels{
			traceIDKey: sc.TraceID().String(),
		})

		return
	}

	m.searchesLatency.Observe(dur.Seconds())
}

func (m *Metrics) IncInsertsTotal() {
	m.insertsTotal.Inc()
}

func (m *Metrics) IncInsertsFailed() {
	m.insertsFailed.Inc()
}

func (m *Metrics) ObserveInsertLatency(ctx context.Context, dur time.Duration) {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		m.insertsLatency.(prometheus.ExemplarObserver).ObserveWithExemplar(dur.Seconds(), prometheus.Labels{
			traceIDKey: sc.TraceID().String(),
		})

		return
	}

	m.insertsLatency.Observe(dur.Seconds())
}

func (m *Metrics) IncDeletesTotal() {
	m.deletesTotal.Inc()
}

func (m *Metrics) IncDeletesFailed() {
	m.deletesFailed.Inc()
}

func (m *Metrics) ObserveDeleteLatency(ctx context.Context, dur time.Duration) {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		m.deletesLatency.(prometheus.ExemplarObserver).ObserveWithExemplar(dur.Seconds(), prometheus.Labels{
			traceIDKey: sc.TraceID().String(),
		})

		return
	}

	m.deletesLatency.Observe(dur.Seconds())
}

func (m *Metrics) Registry() (reg *prometheus.Registry, err error) {
	reg = prometheus.NewRegistry()

	for _, metric := range []prometheus.Collector{
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{
			ReportErrors: false,
		}),
		m.searchesTotal, m.searchesFailed, m.searchesLatency,
		m.insertsTotal, m.insertsFailed, m.insertsLatency,
		m.deletesTotal, m.deletesFailed, m.deletesLatency,
	} {
		if err = reg.Register(metric); err != nil {
			return nil, err
		}
	}

	return reg, nil
}

func (m *Metrics) Shutdown(ctx context.Context) error {
	if m.server == nil {
		return nil
	}

	return m.server.Shutdown(ctx)
}

func newProm() *Metrics {
	return &Metrics{
		searchesTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "searches_received_total",
			Help: "Count of the search requests received by the index",
		}),
		searchesFailed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "searches_failed_total",
			Help: "Count of the failed search requests",
		}),
		searchesLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "search_handling_latency_seconds",
			Help:    "Histogram of search request handling latencies",
			Buckets: []float64{.00001, .00005, .0001, .0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}),

		insertsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "inserts_received_total",
			Help: "Count of the insert requests received by the index",
		}),
		insertsFailed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "inserts_failed_total",
			Help: "Count of the failed insert requests",
		}),
		insertsLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "insert_handling_latency_seconds",
			Help:    "Histogram of insert request handling latencies",
			Buckets: []float64{.00001, .00005, .0001, .0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}),

		deletesTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "deletes_received_total",
			Help: "Count of the delete requests received by the index",
		}),
		deletesFailed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "deletes_failed_total",
			Help: "Count of the failed delete requests",
		}),
		deletesLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "delete_handling_latency_seconds",
			Help:    "Histogram of delete request handling latencies",
			Buckets: []float64{.00001, .00005, .0001, .0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}),
	}
}
