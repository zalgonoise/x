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

// TODO: Otel metrics instead of prom metrics
type Metrics struct {
	// CA metrics
	servicesRegisteredTotal        prometheus.Counter
	servicesRegisteredFailed       prometheus.Counter
	servicesRegistryLatencySeconds prometheus.Histogram

	servicesDeletedTotal          prometheus.Counter
	servicesDeletedFailed         prometheus.Counter
	servicesDeletedLatencySeconds prometheus.Histogram

	certificatesCreatedTotal          *prometheus.CounterVec
	certificatesCreatedFailed         *prometheus.CounterVec
	certificatesCreatedLatencySeconds *prometheus.HistogramVec

	certificatesListedTotal          *prometheus.CounterVec
	certificatesListedFailed         *prometheus.CounterVec
	certificatesListedLatencySeconds *prometheus.HistogramVec

	certificatesDeletedTotal          *prometheus.CounterVec
	certificatesDeletedFailed         *prometheus.CounterVec
	certificatesDeletedLatencySeconds *prometheus.HistogramVec

	certificatesVerifiedTotal          *prometheus.CounterVec
	certificatesVerifiedFailed         *prometheus.CounterVec
	certificatesVerifiedLatencySeconds *prometheus.HistogramVec

	rootCertificateRequestsTotal          prometheus.Counter
	rootCertificateRequestsFailed         prometheus.Counter
	rootCertificateRequestsLatencySeconds prometheus.Histogram

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

		certificatesCreatedTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "certificates_created_total",
			Help: "Count of service certificate creation requests",
		}, []string{"service"}),
		certificatesCreatedFailed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "certificates_created_failed",
			Help: "Count of service certificate creation requests that failed",
		}, []string{"service"}),
		certificatesCreatedLatencySeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "certificates_created_latency_seconds",
			Help:    "Histogram of service certificate creation processing times",
			Buckets: []float64{.00001, .00005, .0001, .0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}, []string{"service"}),

		certificatesListedTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "certificates_listed_total",
			Help: "Count of service certificate listings requested",
		}, []string{"service"}),
		certificatesListedFailed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "certificates_listed_failed",
			Help: "Count of service certificate listings that failed",
		}, []string{"service"}),
		certificatesListedLatencySeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "certificates_listed_latency_seconds",
			Help:    "Histogram of service certificate listing processing times",
			Buckets: []float64{.00001, .00005, .0001, .0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}, []string{"service"}),

		certificatesDeletedTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "certificates_deleted_total",
			Help: "Count of service certificate deletion requests",
		}, []string{"service"}),
		certificatesDeletedFailed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "certificates_deleted_failed",
			Help: "Count of service certificate deletion requests that failed",
		}, []string{"service"}),
		certificatesDeletedLatencySeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "certificates_deleted_latency_seconds",
			Help:    "Histogram of service certificate deletion processing times",
			Buckets: []float64{.00001, .00005, .0001, .0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}, []string{"service"}),

		certificatesVerifiedTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "certificates_verified_total",
			Help: "Count of services' certificates verified by this CA",
		}, []string{"service"}),
		certificatesVerifiedFailed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "certificates_verified_failed",
			Help: "Count of services' certificates verification requests that failed",
		}, []string{"service"}),
		certificatesVerifiedLatencySeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "certificates_verified_latency_seconds",
			Help:    "Histogram of service certificate verification processing times",
			Buckets: []float64{.00001, .00005, .0001, .0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}, []string{"service"}),

		rootCertificateRequestsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "root_certificate_requests_total",
			Help: "Count of root certificate requests",
		}),
		rootCertificateRequestsFailed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "root_certificate_requests_failed",
			Help: "Count of root certificate requests that failed",
		}),
		rootCertificateRequestsLatencySeconds: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "root_certificate_requests_latency_seconds",
			Help:    "Histogram of root certificate request processing times",
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

func (m *Metrics) IncCertificatesCreated(service string) {
	m.certificatesCreatedTotal.WithLabelValues(service).Inc()
}

func (m *Metrics) IncCertificatesCreateFailed(service string) {
	m.certificatesCreatedFailed.WithLabelValues(service).Inc()
}

func (m *Metrics) ObserveCertificatesCreateLatency(ctx context.Context, service string, duration time.Duration) {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		if eo, ok := m.certificatesCreatedLatencySeconds.WithLabelValues(service).(prometheus.ExemplarObserver); ok {
			eo.ObserveWithExemplar(duration.Seconds(), prometheus.Labels{
				traceIDKey: sc.TraceID().String(),
			})

			return
		}
	}

	m.certificatesCreatedLatencySeconds.WithLabelValues(service).Observe(duration.Seconds())
}

func (m *Metrics) IncCertificatesListed(service string) {
	m.certificatesListedTotal.WithLabelValues(service).Inc()
}

func (m *Metrics) IncCertificatesListFailed(service string) {
	m.certificatesListedFailed.WithLabelValues(service).Inc()
}

func (m *Metrics) ObserveCertificatesListLatency(ctx context.Context, service string, duration time.Duration) {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		if eo, ok := m.certificatesListedLatencySeconds.WithLabelValues(service).(prometheus.ExemplarObserver); ok {
			eo.ObserveWithExemplar(duration.Seconds(), prometheus.Labels{
				traceIDKey: sc.TraceID().String(),
			})

			return
		}
	}

	m.certificatesListedLatencySeconds.WithLabelValues(service).Observe(duration.Seconds())
}

func (m *Metrics) IncCertificatesDeleted(service string) {
	m.certificatesDeletedTotal.WithLabelValues(service).Inc()
}

func (m *Metrics) IncCertificatesDeleteFailed(service string) {
	m.certificatesDeletedFailed.WithLabelValues(service).Inc()
}

func (m *Metrics) ObserveCertificatesDeleteLatency(ctx context.Context, service string, duration time.Duration) {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		if eo, ok := m.certificatesDeletedLatencySeconds.WithLabelValues(service).(prometheus.ExemplarObserver); ok {
			eo.ObserveWithExemplar(duration.Seconds(), prometheus.Labels{
				traceIDKey: sc.TraceID().String(),
			})

			return
		}
	}

	m.certificatesDeletedLatencySeconds.WithLabelValues(service).Observe(duration.Seconds())
}

func (m *Metrics) IncCertificatesVerified(service string) {
	m.certificatesVerifiedTotal.WithLabelValues(service).Inc()
}

func (m *Metrics) IncCertificateVerificationFailed(service string) {
	m.certificatesVerifiedFailed.WithLabelValues(service).Inc()
}

func (m *Metrics) ObserveCertificateVerificationLatency(ctx context.Context, service string, duration time.Duration) {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		if eo, ok := m.certificatesVerifiedLatencySeconds.WithLabelValues(service).(prometheus.ExemplarObserver); ok {
			eo.ObserveWithExemplar(duration.Seconds(), prometheus.Labels{
				traceIDKey: sc.TraceID().String(),
			})

			return
		}
	}

	m.certificatesVerifiedLatencySeconds.WithLabelValues(service).Observe(duration.Seconds())
}

func (m *Metrics) IncRootCertificateRequests() {
	m.rootCertificateRequestsTotal.Inc()
}

func (m *Metrics) IncRootCertificateRequestFailed() {
	m.rootCertificateRequestsFailed.Inc()
}

func (m *Metrics) ObserveRootCertificateRequestLatency(ctx context.Context, duration time.Duration) {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		if eo, ok := m.rootCertificateRequestsLatencySeconds.(prometheus.ExemplarObserver); ok {
			eo.ObserveWithExemplar(duration.Seconds(), prometheus.Labels{
				traceIDKey: sc.TraceID().String(),
			})

			return
		}
	}

	m.rootCertificateRequestsLatencySeconds.Observe(duration.Seconds())
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
		m.servicesDeletedTotal,
		m.servicesDeletedFailed,
		m.servicesDeletedLatencySeconds,

		m.certificatesCreatedTotal,
		m.certificatesCreatedFailed,
		m.certificatesCreatedLatencySeconds,
		m.certificatesListedTotal,
		m.certificatesListedFailed,
		m.certificatesListedLatencySeconds,
		m.certificatesDeletedTotal,
		m.certificatesDeletedFailed,
		m.certificatesDeletedLatencySeconds,
		m.certificatesVerifiedTotal,
		m.certificatesVerifiedFailed,
		m.certificatesVerifiedLatencySeconds,

		m.rootCertificateRequestsTotal,
		m.rootCertificateRequestsFailed,
		m.rootCertificateRequestsLatencySeconds,

		m.serviceTokenRequestsTotal,
		m.serviceTokenRequestsFailed,
		m.serviceTokenRequestsLatencySeconds,
		m.serviceTokenVerifyTotal,
		m.serviceTokenVerifyFailed,
		m.serviceTokenVerifyLatencySeconds,

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
