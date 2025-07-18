package ca

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/errs"
	"github.com/zalgonoise/x/reg"

	"github.com/zalgonoise/x/authz/internal/certs"
	"github.com/zalgonoise/x/authz/internal/keygen"
	pb "github.com/zalgonoise/x/authz/pb/authz/v1"
)

const certificateLimit = 2

const (
	errDomain = errs.Domain("x/authz/ca")

	ErrNil     = errs.Kind("nil")
	ErrInvalid = errs.Kind("invalid")

	ErrPublicKey   = errs.Entity("public key")
	ErrPrivateKey  = errs.Entity("private key")
	ErrCertificate = errs.Entity("certificate")
	ErrRepository  = errs.Entity("repository")
)

var (
	ErrNilRepository      = errs.WithDomain(errDomain, ErrNil, ErrRepository)
	ErrNilPrivateKey      = errs.WithDomain(errDomain, ErrNil, ErrPrivateKey)
	ErrInvalidPublicKey   = errs.WithDomain(errDomain, ErrInvalid, ErrPublicKey)
	ErrInvalidCertificate = errs.WithDomain(errDomain, ErrInvalid, ErrCertificate)
)

type Repository interface {
	GetService(ctx context.Context, service string) (pubKey []byte, err error)
	CreateService(ctx context.Context, service string, pubKey []byte) (err error)
	DeleteService(ctx context.Context, service string) error

	ListCertificates(ctx context.Context, service string) (certs []*pb.CertificateResponse, err error)
	CreateCertificate(ctx context.Context, service string, cert []byte, expiry time.Time) error
	DeleteCertificate(ctx context.Context, service string, cert []byte) error

	Shutdown(ctx context.Context) error
}

type Metrics interface {
	IncServiceRegistries()
	IncServiceRegistryFailed()
	ObserveServiceRegistryLatency(ctx context.Context, duration time.Duration)
	IncServiceDeletions()
	IncServiceDeletionFailed()
	ObserveServiceDeletionLatency(ctx context.Context, duration time.Duration)
	IncCertificatesCreated(service string)
	IncCertificatesCreateFailed(service string)
	ObserveCertificatesCreateLatency(ctx context.Context, service string, duration time.Duration)
	IncCertificatesListed(service string)
	IncCertificatesListFailed(service string)
	ObserveCertificatesListLatency(ctx context.Context, service string, duration time.Duration)
	IncCertificatesDeleted(service string)
	IncCertificatesDeleteFailed(service string)
	ObserveCertificatesDeleteLatency(ctx context.Context, service string, duration time.Duration)
	IncCertificatesVerified(service string)
	IncCertificateVerificationFailed(service string)
	ObserveCertificateVerificationLatency(ctx context.Context, service string, duration time.Duration)
	IncRootCertificateRequests()
	IncRootCertificateRequestFailed()
	ObserveRootCertificateRequestLatency(ctx context.Context, duration time.Duration)
}

type CertificateAuthority struct {
	pb.UnimplementedCertificateAuthorityServer

	privateKey *ecdsa.PrivateKey
	ca         *x509.Certificate
	raw        []byte
	durMonth   int

	repository Repository

	r       *reg.Registrar
	logger  *slog.Logger
	tracer  trace.Tracer
	metrics Metrics
}

func NewCertificateAuthority(
	repo Repository,
	privateKey *ecdsa.PrivateKey,
	opts ...cfg.Option[Config],
) (*CertificateAuthority, error) {
	if repo == nil {
		return nil, ErrNilRepository
	}

	if privateKey == nil {
		return nil, ErrNilPrivateKey
	}

	config := cfg.Set(defaultConfig(), opts...)

	template := cfg.Set(certs.DefaultTemplate(), config.template...)
	if template.PrivateKey == nil {
		template.PrivateKey = privateKey
	}

	cert, err := certs.NewCACertificate(template)
	if err != nil {
		return nil, err
	}

	ca, err := certs.Decode(cert)
	if err != nil {
		return nil, err
	}

	logger := slog.New(config.logHandler)

	return &CertificateAuthority{
		privateKey: privateKey,
		ca:         ca,
		raw:        cert,
		durMonth:   template.DurMonth,
		repository: repo,
		r: reg.New(logger,
			[]attribute.KeyValue{attribute.Int("dur_months", template.DurMonth)},
			[]any{slog.Int("dur_months", template.DurMonth)}),
		logger:  logger,
		tracer:  config.tracer,
		metrics: config.metrics,
	}, nil
}

func (ca *CertificateAuthority) Shutdown(ctx context.Context) error {
	return ca.repository.Shutdown(ctx)
}

func (ca *CertificateAuthority) validatePublicKeys(ctx context.Context, service string, key []byte) error {
	storedPub, err := ca.repository.GetService(ctx, service)
	if err != nil {
		return err
	}

	pub, err := keygen.DecodePublic(storedPub)
	if err != nil {
		return err
	}

	reqPub, err := keygen.DecodePublic(key)
	if err != nil {
		return err
	}

	if !pub.Equal(reqPub) {
		return ErrInvalidPublicKey
	}

	return nil
}

func recordError(
	ctx context.Context, err error, message string,
	span trace.Span, traceAttrs []attribute.KeyValue,
	logger *slog.Logger, level slog.Level, logAttrs []any,
) {
	traceAttrs = append(traceAttrs, attribute.String("error", err.Error()))
	logAttrs = append(logAttrs, attribute.String("error", err.Error()))

	span.SetStatus(otelcodes.Error, err.Error())
	span.RecordError(err)
	span.AddEvent(message, trace.WithAttributes(traceAttrs...))

	logger.Log(ctx, level, message, logAttrs...)
}
