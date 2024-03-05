package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"go.opentelemetry.io/otel/trace"
)

const (
	disconnected = 0.0
	connected    = 1.0

	// traceIDKey is used as the trace ID key value in the prometheus.Labels in a prometheus.Exemplar.
	//
	// Its value of `trace_id` complies with the OpenTelemetry specification for metrics' exemplars, as seen in:
	// https://opentelemetry.io/docs/specs/otel/metrics/data-model/#exemplars
	traceIDKey = "trace_id"
)

type Metrics struct {
	// CA metrics
	servicesRegisteredTotal        prometheus.Counter
	servicesRegisteredFailed       prometheus.Counter
	servicesRegistryLatencySeconds prometheus.Histogram

	servicesCertsFetchedTotal          *prometheus.CounterVec
	servicesCertsFetchedFailed         *prometheus.CounterVec
	servicesCertsFetchedLatencySeconds *prometheus.HistogramVec

	servicesCertsVerifiedTotal          *prometheus.CounterVec
	servicesCertsVerifiedFailed         *prometheus.CounterVec
	servicesCertsVerifiedLatencySeconds *prometheus.HistogramVec

	servicesDeletedTotal          prometheus.Counter
	servicesDeletedFailed         prometheus.Counter
	servicesDeletedLatencySeconds prometheus.Histogram

	publicKeyRequestsTotal          prometheus.Counter
	publicKeyRequestsFailed         prometheus.Counter
	publicKeyRequestsLatencySeconds prometheus.Histogram

	// Authz metrics
	serviceLoginRequestsTotal          *prometheus.CounterVec
	serviceLoginRequestsFailed         *prometheus.CounterVec
	serviceLoginRequestsLatencySeconds *prometheus.HistogramVec

	serviceTokenRequestsTotal          *prometheus.CounterVec
	serviceTokenRequestsFailed         *prometheus.CounterVec
	serviceTokenRequestsLatencySeconds *prometheus.HistogramVec

	serviceTokenVerifyTotal          *prometheus.CounterVec
	serviceTokenVerifyFailed         *prometheus.CounterVec
	serviceTokenVerifyLatencySeconds *prometheus.HistogramVec

	// Cron metrics
	schedulerNextTotal         prometheus.Counter
	executorExecTotal          *prometheus.CounterVec
	executorExecFailed         *prometheus.CounterVec
	executorExecLatencySeconds *prometheus.HistogramVec
	executorNextTotal          *prometheus.CounterVec
	selectorSelectTotal        prometheus.Counter
	selectorSelectFailed       prometheus.Counter

	// Third party metrics
	collectors []prometheus.Collector
}

func NewMetrics() *Metrics {
	return &Metrics{
		servicesRegisteredTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "services_registered_total",
			Help: "Count of services registered in this CA",
		}),
		servicesRegisteredFailed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "services_registered_failed",
			Help: "Count of service registry requests that failed",
		}),
		servicesRegistryLatencySeconds: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "services_registered_latency_seconds",
			Help:    "Histogram of service registry processing times",
			Buckets: []float64{.00001, .00005, .0001, .0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}),

		servicesCertsFetchedTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "services_certs_fetched_total",
			Help: "Count of service certificates requested",
		}, []string{"service"}),
		servicesCertsFetchedFailed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "services_certs_fetched_failed",
			Help: "Count of service certificate requests that failed",
		}, []string{"service"}),
		servicesCertsFetchedLatencySeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "services_certs_fetched_latency_seconds",
			Help:    "Histogram of service certificate request processing times",
			Buckets: []float64{.00001, .00005, .0001, .0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}, []string{"service"}),
		servicesCertsVerifiedTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "services_certs_verified_total",
			Help: "Count of services' certificates verified by this CA",
		}, []string{"service"}),
		servicesCertsVerifiedFailed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "services_certs_verified_failed",
			Help: "Count of services' certificates verification requests that failed",
		}, []string{"service"}),
		servicesCertsVerifiedLatencySeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "services_certs_verified_latency_seconds",
			Help:    "Histogram of service certificate verification processing times",
			Buckets: []float64{.00001, .00005, .0001, .0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}, []string{"service"}),
		servicesDeletedTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "services_deleted_total",
			Help: "Count of services deleted from this CA",
		}),
		servicesDeletedFailed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "services_deleted_failed",
			Help: "Count of service deletion requests that failed",
		}),
		servicesDeletedLatencySeconds: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "services_deleted_latency_seconds",
			Help:    "Histogram of service deletion processing times",
			Buckets: []float64{.00001, .00005, .0001, .0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}),
		publicKeyRequestsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "public_key_requests_total",
			Help: "Count of CA public key requests",
		}),
		publicKeyRequestsFailed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "public_key_requests_failed",
			Help: "Count of CA public key requests that failed",
		}),
		publicKeyRequestsLatencySeconds: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "public_key_requests_latency_seconds",
			Help:    "Histogram of CA public key request processing times",
			Buckets: []float64{.00001, .00005, .0001, .0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}),
		serviceLoginRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "service_login_requests_total",
			Help: "Count of Authz service login requests",
		}, []string{"service"}),
		serviceLoginRequestsFailed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "service_login_requests_failed",
			Help: "Count of Authz service login requests that failed",
		}, []string{"service"}),
		serviceLoginRequestsLatencySeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "service_login_requests_latency_seconds",
			Help:    "Histogram of Authz service login requests processing times",
			Buckets: []float64{.00001, .00005, .0001, .0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}, []string{"service"}),
		serviceTokenRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "service_token_requests_total",
			Help: "Count of Authz service token requests",
		}, []string{"service"}),
		serviceTokenRequestsFailed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "service_token_requests_failed",
			Help: "Count of Authz service token requests that failed",
		}, []string{"service"}),
		serviceTokenRequestsLatencySeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "service_token_requests_latency_seconds",
			Help:    "Histogram of Authz service token requests processing times",
			Buckets: []float64{.00001, .00005, .0001, .0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}, []string{"service"}),
		serviceTokenVerifyTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "service_token_verify_total",
			Help: "Count of Authz service token verification requests",
		}, []string{"service"}),
		serviceTokenVerifyFailed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "service_token_verify_failed",
			Help: "Count of Authz service token verification requests that failed",
		}, []string{"service"}),
		serviceTokenVerifyLatencySeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "service_token_verify_latency_seconds",
			Help:    "Histogram of Authz service token verification requests processing times",
			Buckets: []float64{.00001, .00005, .0001, .0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}, []string{"service"}),
		schedulerNextTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cron_scheduler_next_total",
			Help: "Count of cron's scheduler Next calls",
		}),
		executorExecTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "cron_executor_exec_total",
			Help: "Count of cron's executor Exec calls",
		}, []string{"id"}),
		executorExecFailed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "cron_executor_exec_failed",
			Help: "Count of failed cron's executor Exec calls",
		}, []string{"id"}),
		executorExecLatencySeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "cron_executor_exec_latency_seconds",
			Help:    "Histogram of cron's executor Exec calls processing times",
			Buckets: []float64{.00001, .00005, .0001, .0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}, []string{"id"}),
		executorNextTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "cron_executor_next_total",
			Help: "Count of cron's executor Next calls",
		}, []string{"id"}),
		selectorSelectTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cron_selector_select_total",
			Help: "Count of cron's selector Select calls",
		}),
		selectorSelectFailed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cron_selector_select_failed",
			Help: "Count of failed cron's selector Select calls",
		}),
	}
}

func (m *Metrics) IncServiceRegistries() {
	m.servicesRegisteredTotal.Inc()
}

func (m *Metrics) IncServiceRegistryFailed() {
	m.servicesRegisteredFailed.Inc()
}

func (m *Metrics) ObserveServiceRegistryLatency(ctx context.Context, duration time.Duration) {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		if eo, ok := m.servicesRegistryLatencySeconds.(prometheus.ExemplarObserver); ok {
			eo.ObserveWithExemplar(duration.Seconds(), prometheus.Labels{
				traceIDKey: sc.TraceID().String(),
			})

			return
		}
	}

	m.servicesRegistryLatencySeconds.Observe(duration.Seconds())
}

func (m *Metrics) IncServiceCertsFetched(service string) {
	m.servicesCertsFetchedTotal.WithLabelValues(service).Inc()
}

func (m *Metrics) IncServiceCertsFetchFailed(service string) {
	m.servicesCertsFetchedFailed.WithLabelValues(service).Inc()
}

func (m *Metrics) ObserveServiceCertsFetchLatency(ctx context.Context, service string, duration time.Duration) {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		if eo, ok := m.servicesCertsFetchedLatencySeconds.WithLabelValues(service).(prometheus.ExemplarObserver); ok {
			eo.ObserveWithExemplar(duration.Seconds(), prometheus.Labels{
				traceIDKey: sc.TraceID().String(),
			})

			return
		}
	}

	m.servicesCertsFetchedLatencySeconds.WithLabelValues(service).Observe(duration.Seconds())
}

func (m *Metrics) IncVerificationRequests(service string) {
	m.servicesCertsVerifiedTotal.WithLabelValues(service).Inc()
}

func (m *Metrics) IncVerificationFailed(service string) {
	m.servicesCertsVerifiedFailed.WithLabelValues(service).Inc()
}

func (m *Metrics) ObserveVerificationLatency(ctx context.Context, service string, duration time.Duration) {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		if eo, ok := m.servicesCertsVerifiedLatencySeconds.WithLabelValues(service).(prometheus.ExemplarObserver); ok {
			eo.ObserveWithExemplar(duration.Seconds(), prometheus.Labels{
				traceIDKey: sc.TraceID().String(),
			})

			return
		}
	}

	m.servicesCertsVerifiedLatencySeconds.WithLabelValues(service).Observe(duration.Seconds())
}

func (m *Metrics) IncServiceDeletions() {
	m.servicesDeletedTotal.Inc()
}

func (m *Metrics) IncServiceDeletionFailed() {
	m.servicesDeletedFailed.Inc()
}

func (m *Metrics) ObserveServiceDeletionLatency(ctx context.Context, duration time.Duration) {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		if eo, ok := m.servicesDeletedLatencySeconds.(prometheus.ExemplarObserver); ok {
			eo.ObserveWithExemplar(duration.Seconds(), prometheus.Labels{
				traceIDKey: sc.TraceID().String(),
			})

			return
		}
	}

	m.servicesDeletedLatencySeconds.Observe(duration.Seconds())
}

func (m *Metrics) IncPubKeyRequests() {
	m.publicKeyRequestsTotal.Inc()
}

func (m *Metrics) IncPubKeyRequestFailed() {
	m.publicKeyRequestsFailed.Inc()
}

func (m *Metrics) ObservePubKeyRequestLatency(ctx context.Context, duration time.Duration) {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		if eo, ok := m.publicKeyRequestsLatencySeconds.(prometheus.ExemplarObserver); ok {
			eo.ObserveWithExemplar(duration.Seconds(), prometheus.Labels{
				traceIDKey: sc.TraceID().String(),
			})

			return
		}
	}

	m.publicKeyRequestsLatencySeconds.Observe(duration.Seconds())
}

func (m *Metrics) IncServiceLoginRequests(service string) {
	m.serviceLoginRequestsTotal.WithLabelValues(service).Inc()
}

func (m *Metrics) IncServiceLoginFailed(service string) {
	m.serviceLoginRequestsFailed.WithLabelValues(service).Inc()
}

func (m *Metrics) ObserveServiceLoginLatency(ctx context.Context, service string, duration time.Duration) {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		if eo, ok := m.serviceLoginRequestsLatencySeconds.WithLabelValues(service).(prometheus.ExemplarObserver); ok {
			eo.ObserveWithExemplar(duration.Seconds(), prometheus.Labels{
				traceIDKey: sc.TraceID().String(),
			})

			return
		}
	}

	m.serviceLoginRequestsLatencySeconds.WithLabelValues(service).Observe(duration.Seconds())
}

func (m *Metrics) IncServiceTokenRequests(service string) {
	m.serviceTokenRequestsTotal.WithLabelValues(service).Inc()
}

func (m *Metrics) IncServiceTokenFailed(service string) {
	m.serviceTokenRequestsFailed.WithLabelValues(service).Inc()
}

func (m *Metrics) ObserveServiceTokenLatency(ctx context.Context, service string, duration time.Duration) {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		if eo, ok := m.serviceTokenRequestsLatencySeconds.WithLabelValues(service).(prometheus.ExemplarObserver); ok {
			eo.ObserveWithExemplar(duration.Seconds(), prometheus.Labels{
				traceIDKey: sc.TraceID().String(),
			})

			return
		}
	}

	m.serviceTokenRequestsLatencySeconds.WithLabelValues(service).Observe(duration.Seconds())
}

func (m *Metrics) IncServiceTokenVerifications(service string) {
	m.serviceTokenVerifyTotal.WithLabelValues(service).Inc()
}

func (m *Metrics) IncServiceTokenVerificationFailed(service string) {
	m.serviceTokenVerifyFailed.WithLabelValues(service).Inc()
}

func (m *Metrics) ObserveServiceTokenVerificationLatency(ctx context.Context, service string, duration time.Duration) {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		if eo, ok := m.serviceTokenVerifyLatencySeconds.WithLabelValues(service).(prometheus.ExemplarObserver); ok {
			eo.ObserveWithExemplar(duration.Seconds(), prometheus.Labels{
				traceIDKey: sc.TraceID().String(),
			})

			return
		}
	}

	m.serviceTokenVerifyLatencySeconds.WithLabelValues(service).Observe(duration.Seconds())
}

func (m *Metrics) IncSchedulerNextCalls() {
	m.schedulerNextTotal.Inc()
}
func (m *Metrics) IncExecutorExecCalls(id string) {
	m.executorExecTotal.WithLabelValues(id).Inc()
}
func (m *Metrics) IncExecutorExecErrors(id string) {
	m.executorExecFailed.WithLabelValues(id).Inc()
}
func (m *Metrics) ObserveExecLatency(ctx context.Context, id string, dur time.Duration) {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		if eo, ok := m.executorExecLatencySeconds.WithLabelValues(id).(prometheus.ExemplarObserver); ok {
			eo.ObserveWithExemplar(dur.Seconds(), prometheus.Labels{
				traceIDKey: sc.TraceID().String(),
			})

			return
		}
	}

	m.executorExecLatencySeconds.WithLabelValues(id).Observe(dur.Seconds())

}
func (m *Metrics) IncExecutorNextCalls(id string) {
	m.executorNextTotal.WithLabelValues(id).Inc()
}
func (m *Metrics) IncSelectorSelectCalls() {
	m.selectorSelectTotal.Inc()
}
func (m *Metrics) IncSelectorSelectErrors() {
	m.selectorSelectFailed.Inc()
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
		m.servicesRegisteredTotal,
		m.servicesRegisteredFailed,
		m.servicesRegistryLatencySeconds,
		m.servicesCertsFetchedTotal,
		m.servicesCertsFetchedFailed,
		m.servicesCertsFetchedLatencySeconds,
		m.servicesCertsVerifiedTotal,
		m.servicesCertsVerifiedFailed,
		m.servicesCertsVerifiedLatencySeconds,
		m.servicesDeletedTotal,
		m.servicesDeletedFailed,
		m.servicesDeletedLatencySeconds,
		m.serviceTokenRequestsTotal,
		m.serviceTokenRequestsFailed,
		m.serviceTokenRequestsLatencySeconds,
		m.serviceTokenVerifyTotal,
		m.serviceTokenVerifyFailed,
		m.serviceTokenVerifyLatencySeconds,
		m.publicKeyRequestsTotal,
		m.publicKeyRequestsFailed,
		m.publicKeyRequestsLatencySeconds,
		m.schedulerNextTotal,
		m.executorExecTotal,
		m.executorExecFailed,
		m.executorExecLatencySeconds,
		m.executorNextTotal,
		m.selectorSelectTotal,
		m.selectorSelectFailed,
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
