package ca

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"log/slog"
	"slices"
	"time"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/authz/keygen"
	"github.com/zalgonoise/x/authz/log"
	"github.com/zalgonoise/x/authz/metrics"
	pb "github.com/zalgonoise/x/authz/pb/authz/v1"
	"github.com/zalgonoise/x/authz/repository"
	"github.com/zalgonoise/x/errs"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
	Get(ctx context.Context, service string) (pubKey []byte, cert []byte, err error)
	Create(ctx context.Context, service string, pubKey []byte, cert []byte) (err error)
	Delete(ctx context.Context, service string) error
}

type Metrics interface {
	IncServiceRegistries()
	IncServiceRegistryFailed()
	ObserveServiceRegistryLatency(ctx context.Context, duration time.Duration)
	IncServiceCertsFetched(service string)
	IncServiceCertsFetchFailed(service string)
	ObserveServiceCertsFetchLatency(ctx context.Context, service string, duration time.Duration)
	IncServiceDeletions()
	IncServiceDeletionFailed()
	ObserveServiceDeletionLatency(ctx context.Context, duration time.Duration)
	IncPubKeyRequests()
	IncPubKeyRequestFailed()
	ObservePubKeyRequestLatency(ctx context.Context, duration time.Duration)
}

type CertificateAuthority struct {
	pb.UnimplementedCertificateAuthorityServer

	privateKey *ecdsa.PrivateKey
	ca         *x509.Certificate
	cert       *pem.Block
	durMonth   int

	repository Repository

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

	config := cfg.New(opts...)

	if config.logHandler == nil {
		config.logHandler = log.NoOp().Handler()
	}

	if config.tracer == nil {
		config.tracer = noop.NewTracerProvider().Tracer("x/authz/ca")
	}

	if config.metrics == nil {
		config.metrics = metrics.NoOp()
	}

	template := cfg.Set(newDefaultTemplate(), config.template...)
	if template.PrivateKey == nil {
		template.PrivateKey = privateKey
	}

	ca, cert, err := NewCertificate(template)
	if err != nil {
		return nil, err
	}

	return &CertificateAuthority{
		privateKey: privateKey,
		ca:         ca,
		cert:       cert,
		durMonth:   template.DurMonth,
		repository: repo,
		logger:     slog.New(config.logHandler),
		tracer:     config.tracer,
		metrics:    config.metrics,
	}, nil
}

func (ca *CertificateAuthority) Register(
	ctx context.Context, req *pb.CertificateRequest) (*pb.CertificateResponse, error) {
	ctx, span := ca.tracer.Start(ctx, "CertificateAuthority.Register", trace.WithAttributes(
		attribute.String("service", req.Service),
		attribute.String("pub_key", string(req.PublicKey)),
		attribute.Bool("with_csr", req.SigningRequest != nil),
	))
	defer span.End()

	ca.metrics.IncServiceRegistries()
	ca.logger.DebugContext(ctx, "new registry request",
		slog.String("service", req.Service), slog.Bool("with_csr", req.SigningRequest != nil))

	if err := req.ValidateAll(); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		ca.metrics.IncServiceRegistryFailed()
		ca.logger.WarnContext(ctx, "invalid request",
			slog.String("error", err.Error()), slog.Any("request", req))

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	storedPubKey, storedCert, err := ca.repository.Get(ctx, req.Service)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		ca.metrics.IncServiceRegistryFailed()
		ca.logger.ErrorContext(ctx, "failed to get service from DB",
			slog.String("error", err.Error()), slog.String("service", req.Service))

		return nil, status.Error(codes.Internal, err.Error())
	}

	// service is already registered, just return the stored certificate
	if err == nil && len(storedPubKey) > 0 && len(storedCert) > 0 {
		return &pb.CertificateResponse{
			Certificate: storedCert,
		}, nil
	}

	pubKey, err := keygen.DecodePublic(req.PublicKey)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		ca.metrics.IncServiceRegistryFailed()
		ca.logger.WarnContext(ctx, "invalid request",
			slog.String("error", err.Error()), slog.String("pub_key", string(req.PublicKey)))

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	cert, err := newCertFromCSR(ca.ca.Version, ca.durMonth, toCSR(req.Service, req.SigningRequest))
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		ca.metrics.IncServiceRegistryFailed()
		ca.logger.ErrorContext(ctx, "failed to generate new serial number",
			slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	signedCert, err := newCertificate(cert, ca.ca, pubKey, ca.privateKey)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		ca.metrics.IncServiceRegistryFailed()
		ca.logger.ErrorContext(ctx, "failed to generate new certificate",
			slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := ca.repository.Create(ctx, req.Service, req.PublicKey, signedCert); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		ca.metrics.IncServiceRegistryFailed()
		ca.logger.ErrorContext(ctx, "failed to write certificate to DB",
			slog.String("error", err.Error()), slog.String("certificate", string(signedCert)))

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CertificateResponse{Certificate: signedCert}, nil
}

func (ca *CertificateAuthority) GetCertificate(ctx context.Context, req *pb.CertificateRequest) (*pb.CertificateResponse, error) {
	ctx, span := ca.tracer.Start(ctx, "CertificateAuthority.GetCertificate", trace.WithAttributes(
		attribute.String("service", req.Service),
		attribute.String("pub_key", string(req.PublicKey)),
		attribute.Bool("with_csr", req.SigningRequest != nil),
	))
	defer span.End()

	ca.metrics.IncServiceCertsFetched(req.Service)
	ca.logger.DebugContext(ctx, "new certificate request",
		slog.String("service", req.Service), slog.Bool("with_csr", req.SigningRequest != nil))

	if err := req.ValidateAll(); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		ca.metrics.IncServiceCertsFetchFailed(req.Service)
		ca.logger.WarnContext(ctx, "invalid request",
			slog.String("error", err.Error()), slog.Any("request", req))

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	pubKey, cert, err := ca.repository.Get(ctx, req.Service)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			ca.logger.DebugContext(ctx, "service was not found",
				slog.String("error", err.Error()), slog.String("service", req.Service))

			return nil, status.Error(codes.NotFound, err.Error())
		}

		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		ca.metrics.IncServiceCertsFetchFailed(req.Service)
		ca.logger.WarnContext(ctx, "failed to fetch service from the DB",
			slog.String("error", err.Error()), slog.String("service", req.Service))

		return nil, status.Error(codes.Internal, err.Error())
	}

	if !slices.Equal(pubKey, req.PublicKey) {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		ca.metrics.IncServiceCertsFetchFailed(req.Service)
		ca.logger.WarnContext(ctx, "mismatching public keys")

		return nil, status.Error(codes.PermissionDenied, ErrInvalidPublicKey.Error())
	}

	return &pb.CertificateResponse{Certificate: cert}, nil
}

func (ca *CertificateAuthority) DeleteService(ctx context.Context, req *pb.DeletionRequest) (*pb.DeletionResponse, error) {
	ctx, span := ca.tracer.Start(ctx, "CertificateAuthority.DeleteService", trace.WithAttributes(
		attribute.String("service", req.Service),
		attribute.String("pub_key", string(req.PublicKey)),
		attribute.String("cert", string(req.Certificate)),
	))
	defer span.End()

	ca.metrics.IncServiceDeletions()
	ca.logger.DebugContext(ctx, "service deletion request",
		slog.String("service", req.Service))

	if err := req.ValidateAll(); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		ca.metrics.IncServiceDeletionFailed()
		ca.logger.WarnContext(ctx, "invalid request",
			slog.String("error", err.Error()), slog.Any("request", req))

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	pubKey, cert, err := ca.repository.Get(ctx, req.Service)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			ca.logger.DebugContext(ctx, "service was not found",
				slog.String("error", err.Error()), slog.String("service", req.Service))

			return nil, status.Error(codes.NotFound, err.Error())
		}

		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		ca.metrics.IncServiceDeletionFailed()
		ca.logger.WarnContext(ctx, "failed to fetch service from the DB",
			slog.String("error", err.Error()), slog.String("service", req.Service))

		return nil, status.Error(codes.Internal, err.Error())
	}

	if !slices.Equal(pubKey, []byte(req.PublicKey)) {
		span.SetStatus(otelcodes.Error, ErrInvalidPublicKey.Error())
		span.RecordError(err)
		ca.metrics.IncServiceDeletionFailed()
		ca.logger.WarnContext(ctx, "mismatching public keys")

		return nil, status.Error(codes.PermissionDenied, ErrInvalidPublicKey.Error())
	}

	if !slices.Equal(cert, []byte(req.Certificate)) {
		span.SetStatus(otelcodes.Error, ErrInvalidCertificate.Error())
		span.RecordError(err)
		ca.metrics.IncServiceDeletionFailed()
		ca.logger.WarnContext(ctx, "mismatching certificates")

		return nil, status.Error(codes.PermissionDenied, ErrInvalidCertificate.Error())
	}

	if err = ca.repository.Delete(ctx, req.Service); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		ca.metrics.IncServiceDeletionFailed()
		ca.logger.ErrorContext(ctx, "failed to remove service from DB",
			slog.String("error", err.Error()), slog.String("service", req.Service))

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeletionResponse{}, nil
}

func (ca *CertificateAuthority) PublicKey(ctx context.Context, _ *pb.PublicKeyRequest) (*pb.PublicKeyResponse, error) {
	ctx, span := ca.tracer.Start(ctx, "CertificateAuthority.PublicKeyRequest")
	defer span.End()

	ca.metrics.IncPubKeyRequests()
	ca.logger.DebugContext(ctx, "CA's public key request")

	key, err := keygen.EncodePublic(&ca.privateKey.PublicKey)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		ca.metrics.IncPubKeyRequestFailed()
		ca.logger.ErrorContext(ctx, "failed to marshal CA's public key",
			slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.PublicKeyResponse{PublicKey: key}, nil
}
