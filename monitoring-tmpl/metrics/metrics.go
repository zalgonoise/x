package metrics

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/trace"
)

const traceIDKey = "trace_id" // https://opentelemetry.io/docs/specs/otel/metrics/data-model/#exemplars

type Metrics struct {
	reqReceived prometheus.Counter
	reqFailed   prometheus.Counter
	reqLatency  prometheus.Histogram
}

func (m *Metrics) IncRequestsReceived() {
	m.reqReceived.Inc()
}

func (m *Metrics) IncRequestsFailed() {
	m.reqFailed.Inc()
}

func (m *Metrics) ObserveHandlingLatency(ctx context.Context, dur time.Duration) {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		m.reqLatency.(prometheus.ExemplarObserver).ObserveWithExemplar(dur.Seconds(), prometheus.Labels{
			traceIDKey: sc.TraceID().String(),
		})

		return
	}

	m.reqLatency.Observe(dur.Seconds())
}

func (m *Metrics) Registry() (reg *prometheus.Registry, err error) {
	reg = prometheus.NewRegistry()

	for _, metric := range []prometheus.Collector{
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{
			ReportErrors: false,
		}),
		m.reqReceived,
		m.reqFailed,
		m.reqLatency,
	} {
		if err = reg.Register(metric); err != nil {
			return nil, err
		}
	}

	return reg, nil
}

func NewMetrics() *Metrics {
	return &Metrics{
		reqReceived: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "req_received_total",
			Help: "Count of the requests received by the service handler",
		}),
		reqFailed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "req_failed_total",
			Help: "Count of the requests that have failed, in the service handler",
		}),
		reqLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "req_handling_latency_seconds",
			Help:    "Histogram of request handling latencies",
			Buckets: []float64{.00001, .00005, .0001, .0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}),
	}
}

func NewServer(port int, registry *prometheus.Registry) *http.Server {
	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		Registry:          registry,
		EnableOpenMetrics: true,
	}))

	server := &http.Server{
		Handler:      mux,
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	return server
}
