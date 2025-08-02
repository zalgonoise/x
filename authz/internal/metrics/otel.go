package metrics

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

const defaultInterval = 500 * time.Millisecond
const ServiceName = "authz"

type ShutdownFunc func(ctx context.Context) error

func Meter() metric.Meter {
	return otel.GetMeterProvider().Meter(ServiceName)
}

var bucketBoundaries = []float64{
	.00001, .00005, .0001, .0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10,
}

type Otel struct {
	// CA metrics
	servicesRegisteredTotal        metric.Int64Counter
	servicesRegisteredFailed       metric.Int64Counter
	servicesRegistryLatencySeconds metric.Float64Histogram

	servicesDeletedTotal          metric.Int64Counter
	servicesDeletedFailed         metric.Int64Counter
	servicesDeletedLatencySeconds metric.Float64Histogram

	certificatesCreatedTotal          metric.Int64Counter
	certificatesCreatedFailed         metric.Int64Counter
	certificatesCreatedLatencySeconds metric.Float64Histogram

	certificatesListedTotal          metric.Int64Counter
	certificatesListedFailed         metric.Int64Counter
	certificatesListedLatencySeconds metric.Float64Histogram

	certificatesDeletedTotal          metric.Int64Counter
	certificatesDeletedFailed         metric.Int64Counter
	certificatesDeletedLatencySeconds metric.Float64Histogram

	certificatesVerifiedTotal          metric.Int64Counter
	certificatesVerifiedFailed         metric.Int64Counter
	certificatesVerifiedLatencySeconds metric.Float64Histogram

	rootCertificateRequestsTotal          metric.Int64Counter
	rootCertificateRequestsFailed         metric.Int64Counter
	rootCertificateRequestsLatencySeconds metric.Float64Histogram

	// Authz metrics
	serviceLoginRequestsTotal          metric.Int64Counter
	serviceLoginRequestsFailed         metric.Int64Counter
	serviceLoginRequestsLatencySeconds metric.Float64Histogram

	serviceTokenRequestsTotal          metric.Int64Counter
	serviceTokenRequestsFailed         metric.Int64Counter
	serviceTokenRequestsLatencySeconds metric.Float64Histogram

	serviceTokenVerifyTotal          metric.Int64Counter
	serviceTokenVerifyFailed         metric.Int64Counter
	serviceTokenVerifyLatencySeconds metric.Float64Histogram

	// Cron metrics
	schedulerNextTotal         metric.Int64Counter
	executorExecTotal          metric.Int64Counter
	executorExecFailed         metric.Int64Counter
	executorExecLatencySeconds metric.Float64Histogram
	executorNextTotal          metric.Int64Counter
	selectorSelectTotal        metric.Int64Counter
	selectorSelectFailed       metric.Int64Counter
}

func NewOtel() (*Otel, error) {
	servicesRegisteredTotal, err := Meter().Int64Counter(
		"services_registered_total",
		metric.WithUnit("services"),
		metric.WithDescription("Count of services registered in this CA"),
	)
	if err != nil {
		return nil, err
	}

	servicesRegisteredFailed, err := Meter().Int64Counter(
		"services_registered_failed",
		metric.WithUnit("services"),
		metric.WithDescription("Count of service registry requests that failed"),
	)
	if err != nil {
		return nil, err
	}

	servicesRegistryLatencySeconds, err := Meter().Float64Histogram(
		"services_registered_latency_seconds",
		metric.WithUnit("s"),
		metric.WithDescription("Histogram of service registry processing times"),
		metric.WithExplicitBucketBoundaries(bucketBoundaries...))
	if err != nil {
		return nil, err
	}

	servicesDeletedTotal, err := Meter().Int64Counter(
		"services_deleted_total",
		metric.WithUnit("services"),
		metric.WithDescription("Count of services deleted from this CA"),
	)
	if err != nil {
		return nil, err
	}

	servicesDeletedFailed, err := Meter().Int64Counter(
		"services_deleted_failed",
		metric.WithUnit("services"),
		metric.WithDescription("Count of service deletion requests that failed"),
	)
	if err != nil {
		return nil, err
	}

	servicesDeletedLatencySeconds, err := Meter().Float64Histogram(
		"services_deleted_latency_seconds",
		metric.WithUnit("s"),
		metric.WithDescription("Histogram of service deletion processing times"),
		metric.WithExplicitBucketBoundaries(bucketBoundaries...))
	if err != nil {
		return nil, err
	}

	certificatesCreatedTotal, err := Meter().Int64Counter(
		"certificates_created_total",
		metric.WithUnit("services"),
		metric.WithDescription("Count of service certificate creation requests"),
	)
	if err != nil {
		return nil, err
	}

	certificatesCreatedFailed, err := Meter().Int64Counter(
		"certificates_created_failed",
		metric.WithUnit("services"),
		metric.WithDescription("Count of service certificate creation requests that failed"),
	)
	if err != nil {
		return nil, err
	}

	certificatesCreatedLatencySeconds, err := Meter().Float64Histogram(
		"certificates_created_latency_seconds",
		metric.WithUnit("s"),
		metric.WithDescription("Histogram of service certificate creation processing times"),
		metric.WithExplicitBucketBoundaries(bucketBoundaries...))
	if err != nil {
		return nil, err
	}

	certificatesListedTotal, err := Meter().Int64Counter(
		"certificates_listed_total",
		metric.WithUnit("services"),
		metric.WithDescription("Count of service certificate listings requested"),
	)
	if err != nil {
		return nil, err
	}

	certificatesListedFailed, err := Meter().Int64Counter(
		"certificates_listed_failed",
		metric.WithUnit("services"),
		metric.WithDescription("Count of service certificate listings that failed"),
	)
	if err != nil {
		return nil, err
	}

	certificatesListedLatencySeconds, err := Meter().Float64Histogram(
		"certificates_listed_latency_seconds",
		metric.WithUnit("s"),
		metric.WithDescription("Histogram of service certificate listing processing times"),
		metric.WithExplicitBucketBoundaries(bucketBoundaries...))
	if err != nil {
		return nil, err
	}

	certificatesDeletedTotal, err := Meter().Int64Counter(
		"certificates_deleted_total",
		metric.WithUnit("services"),
		metric.WithDescription("Count of service certificate deletion requests"),
	)
	if err != nil {
		return nil, err
	}

	certificatesDeletedFailed, err := Meter().Int64Counter(
		"certificates_deleted_failed",
		metric.WithUnit("services"),
		metric.WithDescription("Count of service certificate deletion requests that failed"),
	)
	if err != nil {
		return nil, err
	}

	certificatesDeletedLatencySeconds, err := Meter().Float64Histogram(
		"certificates_deleted_latency_seconds",
		metric.WithUnit("s"),
		metric.WithDescription("Histogram of service certificate deletion processing times"),
		metric.WithExplicitBucketBoundaries(bucketBoundaries...))
	if err != nil {
		return nil, err
	}

	certificatesVerifiedTotal, err := Meter().Int64Counter(
		"certificates_verified_total",
		metric.WithUnit("services"),
		metric.WithDescription("Count of services' certificates verified by this CA"),
	)
	if err != nil {
		return nil, err
	}

	certificatesVerifiedFailed, err := Meter().Int64Counter(
		"certificates_verified_failed",
		metric.WithUnit("services"),
		metric.WithDescription("Count of services' certificates verification requests that failed"),
	)
	if err != nil {
		return nil, err
	}

	certificatesVerifiedLatencySeconds, err := Meter().Float64Histogram(
		"certificates_verified_latency_seconds",
		metric.WithUnit("s"),
		metric.WithDescription("Histogram of service certificate verification processing times"),
		metric.WithExplicitBucketBoundaries(bucketBoundaries...))
	if err != nil {
		return nil, err
	}

	rootCertificateRequestsTotal, err := Meter().Int64Counter(
		"root_certificate_requests_total",
		metric.WithUnit("services"),
		metric.WithDescription("Count of root certificate requests"),
	)
	if err != nil {
		return nil, err
	}

	rootCertificateRequestsFailed, err := Meter().Int64Counter(
		"root_certificate_requests_failed",
		metric.WithUnit("services"),
		metric.WithDescription("Count of root certificate requests that failed"),
	)
	if err != nil {
		return nil, err
	}

	rootCertificateRequestsLatencySeconds, err := Meter().Float64Histogram(
		"root_certificate_requests_latency_seconds",
		metric.WithUnit("s"),
		metric.WithDescription("Histogram of root certificate request processing times"),
		metric.WithExplicitBucketBoundaries(bucketBoundaries...))
	if err != nil {
		return nil, err
	}

	serviceLoginRequestsTotal, err := Meter().Int64Counter(
		"service_login_requests_total",
		metric.WithUnit("services"),
		metric.WithDescription("Count of Authz service login requests"),
	)
	if err != nil {
		return nil, err
	}

	serviceLoginRequestsFailed, err := Meter().Int64Counter(
		"service_login_requests_failed",
		metric.WithUnit("services"),
		metric.WithDescription("Count of Authz service login requests that failed"),
	)
	if err != nil {
		return nil, err
	}

	serviceLoginRequestsLatencySeconds, err := Meter().Float64Histogram(
		"service_login_requests_latency_seconds",
		metric.WithUnit("s"),
		metric.WithDescription("Histogram of Authz service login requests processing times"),
		metric.WithExplicitBucketBoundaries(bucketBoundaries...))
	if err != nil {
		return nil, err
	}

	serviceTokenRequestsTotal, err := Meter().Int64Counter(
		"service_token_requests_total",
		metric.WithUnit("services"),
		metric.WithDescription("Count of Authz service token requests"),
	)
	if err != nil {
		return nil, err
	}

	serviceTokenRequestsFailed, err := Meter().Int64Counter(
		"service_token_requests_failed",
		metric.WithUnit("services"),
		metric.WithDescription("Count of Authz service token requests that failed"),
	)
	if err != nil {
		return nil, err
	}

	serviceTokenRequestsLatencySeconds, err := Meter().Float64Histogram(
		"service_token_requests_latency_seconds",
		metric.WithUnit("s"),
		metric.WithDescription("Histogram of Authz service token requests processing times"),
		metric.WithExplicitBucketBoundaries(bucketBoundaries...))
	if err != nil {
		return nil, err
	}

	serviceTokenVerifyTotal, err := Meter().Int64Counter(
		"service_token_verify_total",
		metric.WithUnit("services"),
		metric.WithDescription("Count of Authz service token verification requests"),
	)
	if err != nil {
		return nil, err
	}

	serviceTokenVerifyFailed, err := Meter().Int64Counter(
		"service_token_verify_failed",
		metric.WithUnit("services"),
		metric.WithDescription("Count of Authz service token verification requests that failed"),
	)
	if err != nil {
		return nil, err
	}

	serviceTokenVerifyLatencySeconds, err := Meter().Float64Histogram(
		"service_token_verify_latency_seconds",
		metric.WithUnit("s"),
		metric.WithDescription("Histogram of Authz service token verification requests processing times"),
		metric.WithExplicitBucketBoundaries(bucketBoundaries...))
	if err != nil {
		return nil, err
	}

	schedulerNextTotal, err := Meter().Int64Counter(
		"cron_scheduler_next_total",
		metric.WithUnit("services"),
		metric.WithDescription("Count of cron's scheduler Next calls"),
	)
	if err != nil {
		return nil, err
	}

	executorExecTotal, err := Meter().Int64Counter(
		"cron_executor_exec_total",
		metric.WithUnit("services"),
		metric.WithDescription("Count of cron's executor Exec calls"),
	)
	if err != nil {
		return nil, err
	}

	executorExecFailed, err := Meter().Int64Counter(
		"cron_executor_exec_failed",
		metric.WithUnit("services"),
		metric.WithDescription("Count of failed cron's executor Exec calls"),
	)
	if err != nil {
		return nil, err
	}

	executorExecLatencySeconds, err := Meter().Float64Histogram(
		"cron_executor_exec_latency_seconds",
		metric.WithUnit("s"),
		metric.WithDescription("Histogram of cron's executor Exec calls processing times"),
		metric.WithExplicitBucketBoundaries(bucketBoundaries...))
	if err != nil {
		return nil, err
	}

	executorNextTotal, err := Meter().Int64Counter(
		"cron_executor_next_total",
		metric.WithUnit("services"),
		metric.WithDescription("Count of cron's executor Next calls"),
	)
	if err != nil {
		return nil, err
	}

	selectorSelectTotal, err := Meter().Int64Counter(
		"cron_selector_select_total",
		metric.WithUnit("services"),
		metric.WithDescription("Count of cron's selector Select calls"),
	)
	if err != nil {
		return nil, err
	}

	selectorSelectFailed, err := Meter().Int64Counter(
		"cron_selector_select_failed",
		metric.WithUnit("services"),
		metric.WithDescription("Count of failed cron's selector Select calls"),
	)
	if err != nil {
		return nil, err
	}

	return &Otel{
		servicesRegisteredTotal:               servicesRegisteredTotal,
		servicesRegisteredFailed:              servicesRegisteredFailed,
		servicesRegistryLatencySeconds:        servicesRegistryLatencySeconds,
		servicesDeletedTotal:                  servicesDeletedTotal,
		servicesDeletedFailed:                 servicesDeletedFailed,
		servicesDeletedLatencySeconds:         servicesDeletedLatencySeconds,
		certificatesCreatedTotal:              certificatesCreatedTotal,
		certificatesCreatedFailed:             certificatesCreatedFailed,
		certificatesCreatedLatencySeconds:     certificatesCreatedLatencySeconds,
		certificatesListedTotal:               certificatesListedTotal,
		certificatesListedFailed:              certificatesListedFailed,
		certificatesListedLatencySeconds:      certificatesListedLatencySeconds,
		certificatesDeletedTotal:              certificatesDeletedTotal,
		certificatesDeletedFailed:             certificatesDeletedFailed,
		certificatesDeletedLatencySeconds:     certificatesDeletedLatencySeconds,
		certificatesVerifiedTotal:             certificatesVerifiedTotal,
		certificatesVerifiedFailed:            certificatesVerifiedFailed,
		certificatesVerifiedLatencySeconds:    certificatesVerifiedLatencySeconds,
		rootCertificateRequestsTotal:          rootCertificateRequestsTotal,
		rootCertificateRequestsFailed:         rootCertificateRequestsFailed,
		rootCertificateRequestsLatencySeconds: rootCertificateRequestsLatencySeconds,
		serviceLoginRequestsTotal:             serviceLoginRequestsTotal,
		serviceLoginRequestsFailed:            serviceLoginRequestsFailed,
		serviceLoginRequestsLatencySeconds:    serviceLoginRequestsLatencySeconds,
		serviceTokenRequestsTotal:             serviceTokenRequestsTotal,
		serviceTokenRequestsFailed:            serviceTokenRequestsFailed,
		serviceTokenRequestsLatencySeconds:    serviceTokenRequestsLatencySeconds,
		serviceTokenVerifyTotal:               serviceTokenVerifyTotal,
		serviceTokenVerifyFailed:              serviceTokenVerifyFailed,
		serviceTokenVerifyLatencySeconds:      serviceTokenVerifyLatencySeconds,
		schedulerNextTotal:                    schedulerNextTotal,
		executorExecTotal:                     executorExecTotal,
		executorExecFailed:                    executorExecFailed,
		executorExecLatencySeconds:            executorExecLatencySeconds,
		executorNextTotal:                     executorNextTotal,
		selectorSelectTotal:                   selectorSelectTotal,
		selectorSelectFailed:                  selectorSelectFailed,
	}, nil
}

func (m *Otel) IncServiceRegistries(ctx context.Context) {
	m.servicesRegisteredTotal.Add(ctx, 1)
}

func (m *Otel) IncServiceRegistryFailed(ctx context.Context) {
	m.servicesRegisteredFailed.Add(ctx, 1)
}

func (m *Otel) ObserveServiceRegistryLatency(ctx context.Context, duration time.Duration) {
	m.servicesRegistryLatencySeconds.Record(ctx, duration.Seconds())
}

func (m *Otel) IncServiceDeletions(ctx context.Context) {
	m.servicesDeletedTotal.Add(ctx, 1)
}

func (m *Otel) IncServiceDeletionFailed(ctx context.Context) {
	m.servicesDeletedFailed.Add(ctx, 1)
}

func (m *Otel) ObserveServiceDeletionLatency(ctx context.Context, duration time.Duration) {
	m.servicesDeletedLatencySeconds.Record(ctx, duration.Seconds())
}

func (m *Otel) IncCertificatesCreated(ctx context.Context, service string) {
	m.certificatesCreatedTotal.Add(ctx, 1, metric.WithAttributes(attribute.String("service", service)))
}

func (m *Otel) IncCertificatesCreateFailed(ctx context.Context, service string) {
	m.certificatesCreatedFailed.Add(ctx, 1, metric.WithAttributes(attribute.String("service", service)))
}

func (m *Otel) ObserveCertificatesCreateLatency(ctx context.Context, service string, duration time.Duration) {
	m.certificatesCreatedLatencySeconds.Record(ctx, duration.Seconds(),
		metric.WithAttributes(attribute.String("service", service)))
}

func (m *Otel) IncCertificatesListed(ctx context.Context, service string) {
	m.certificatesListedTotal.Add(ctx, 1, metric.WithAttributes(attribute.String("service", service)))
}

func (m *Otel) IncCertificatesListFailed(ctx context.Context, service string) {
	m.certificatesListedFailed.Add(ctx, 1, metric.WithAttributes(attribute.String("service", service)))
}

func (m *Otel) ObserveCertificatesListLatency(ctx context.Context, service string, duration time.Duration) {
	m.certificatesListedLatencySeconds.Record(ctx, duration.Seconds(),
		metric.WithAttributes(attribute.String("service", service)))
}

func (m *Otel) IncCertificatesDeleted(ctx context.Context, service string) {
	m.certificatesDeletedTotal.Add(ctx, 1, metric.WithAttributes(attribute.String("service", service)))
}

func (m *Otel) IncCertificatesDeleteFailed(ctx context.Context, service string) {
	m.certificatesDeletedFailed.Add(ctx, 1, metric.WithAttributes(attribute.String("service", service)))
}

func (m *Otel) ObserveCertificatesDeleteLatency(ctx context.Context, service string, duration time.Duration) {
	m.certificatesDeletedLatencySeconds.Record(ctx, duration.Seconds(),
		metric.WithAttributes(attribute.String("service", service)))
}

func (m *Otel) IncCertificatesVerified(ctx context.Context, service string) {
	m.certificatesVerifiedTotal.Add(ctx, 1, metric.WithAttributes(attribute.String("service", service)))
}

func (m *Otel) IncCertificateVerificationFailed(ctx context.Context, service string) {
	m.certificatesVerifiedFailed.Add(ctx, 1, metric.WithAttributes(attribute.String("service", service)))
}

func (m *Otel) ObserveCertificateVerificationLatency(ctx context.Context, service string, duration time.Duration) {
	m.certificatesVerifiedLatencySeconds.Record(ctx, duration.Seconds(),
		metric.WithAttributes(attribute.String("service", service)))
}

func (m *Otel) IncRootCertificateRequests(ctx context.Context) {
	m.rootCertificateRequestsTotal.Add(ctx, 1)
}

func (m *Otel) IncRootCertificateRequestFailed(ctx context.Context) {
	m.rootCertificateRequestsFailed.Add(ctx, 1)
}

func (m *Otel) ObserveRootCertificateRequestLatency(ctx context.Context, duration time.Duration) {
	m.rootCertificateRequestsLatencySeconds.Record(ctx, duration.Seconds())
}

func (m *Otel) IncServiceLoginRequests(ctx context.Context, service string) {
	m.serviceLoginRequestsTotal.Add(ctx, 1, metric.WithAttributes(attribute.String("service", service)))
}

func (m *Otel) IncServiceLoginFailed(ctx context.Context, service string) {
	m.serviceLoginRequestsFailed.Add(ctx, 1, metric.WithAttributes(attribute.String("service", service)))
}

func (m *Otel) ObserveServiceLoginLatency(ctx context.Context, service string, duration time.Duration) {
	m.serviceLoginRequestsLatencySeconds.Record(ctx, duration.Seconds(),
		metric.WithAttributes(attribute.String("service", service)))
}

func (m *Otel) IncServiceTokenRequests(ctx context.Context, service string) {
	m.serviceTokenRequestsTotal.Add(ctx, 1, metric.WithAttributes(attribute.String("service", service)))
}

func (m *Otel) IncServiceTokenFailed(ctx context.Context, service string) {
	m.serviceTokenRequestsFailed.Add(ctx, 1, metric.WithAttributes(attribute.String("service", service)))
}

func (m *Otel) ObserveServiceTokenLatency(ctx context.Context, service string, duration time.Duration) {
	m.serviceTokenRequestsLatencySeconds.Record(ctx, duration.Seconds(),
		metric.WithAttributes(attribute.String("service", service)))
}

func (m *Otel) IncServiceTokenVerifications(ctx context.Context, service string) {
	m.serviceTokenVerifyTotal.Add(ctx, 1, metric.WithAttributes(attribute.String("service", service)))
}

func (m *Otel) IncServiceTokenVerificationFailed(ctx context.Context, service string) {
	m.serviceTokenVerifyFailed.Add(ctx, 1, metric.WithAttributes(attribute.String("service", service)))
}

func (m *Otel) ObserveServiceTokenVerificationLatency(ctx context.Context, service string, duration time.Duration) {
	m.serviceTokenVerifyLatencySeconds.Record(ctx, duration.Seconds(),
		metric.WithAttributes(attribute.String("service", service)))
}

func (m *Otel) IncSchedulerNextCalls(ctx context.Context) {
	m.schedulerNextTotal.Add(ctx, 1)
}

func (m *Otel) IncExecutorExecCalls(ctx context.Context, id string) {
	m.executorExecTotal.Add(ctx, 1, metric.WithAttributes(attribute.String("id", id)))
}

func (m *Otel) IncExecutorExecErrors(ctx context.Context, id string) {
	m.executorExecFailed.Add(ctx, 1, metric.WithAttributes(attribute.String("id", id)))
}

func (m *Otel) ObserveExecLatency(ctx context.Context, id string, dur time.Duration) {
	m.executorExecLatencySeconds.Record(ctx, dur.Seconds(), metric.WithAttributes(attribute.String("id", id)))
}

func (m *Otel) IncExecutorNextCalls(ctx context.Context, id string) {
	m.executorNextTotal.Add(ctx, 1, metric.WithAttributes(attribute.String("id", id)))
}

func (m *Otel) IncSelectorSelectCalls(ctx context.Context) {
	m.selectorSelectTotal.Add(ctx, 1)
}

func (m *Otel) IncSelectorSelectErrors(ctx context.Context) {
	m.selectorSelectFailed.Add(ctx, 1)
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
