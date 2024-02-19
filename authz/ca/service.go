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
	Create(ctx context.Context, service string, pubKey []byte, cert []byte, expiry time.Time) (err error)
	Delete(ctx context.Context, service string) error
}

type Metrics interface {
	IncServiceRegistries()
	IncServiceRegistryFailed()
	ObserveServiceRegistryLatency(ctx context.Context, duration time.Duration)
	IncServiceCertsFetched(service string)
	IncServiceCertsFetchFailed(service string)
	ObserveServiceCertsFetchLatency(ctx context.Context, service string, duration time.Duration)
	IncVerificationRequests(service string)
	IncVerificationFailed(service string)
	ObserveVerificationLatency(ctx context.Context, service string, duration time.Duration)
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

	start := time.Now()
	defer func() {
		ca.metrics.ObserveServiceRegistryLatency(ctx, time.Since(start))
	}()

	ca.metrics.IncServiceRegistries()
	ca.logger.DebugContext(ctx, "new registry request",
		slog.String("service", req.Service), slog.Bool("with_csr", req.SigningRequest != nil))

	exit := withExit[pb.CertificateRequest, pb.CertificateResponse](
		ctx, ca, req, ca.metrics.IncServiceRegistryFailed, span,
	)

	if err := req.ValidateAll(); err != nil {
		return exit(codes.InvalidArgument, "invalid request", err)
	}

	storedPubKey, storedCert, err := ca.repository.Get(ctx, req.Service)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return exit(codes.Internal, "failed to get service from DB", err)
	}

	// service is already registered, just return the stored certificate
	if err == nil && len(storedPubKey) > 0 && len(storedCert) > 0 {
		return &pb.CertificateResponse{
			Certificate: storedCert,
		}, nil
	}

	pubKey, err := keygen.DecodePublic(req.PublicKey)
	if err != nil {
		return exit(codes.InvalidArgument, "invalid request", err)
	}

	cert, err := newCertFromCSR(ca.ca.Version, ca.durMonth, toCSR(req.Service, pubKey, req.SigningRequest))
	if err != nil {
		return exit(codes.Internal, "failed to generate new serial number", err)
	}

	signedCert, err := keygen.EncodeCertificate(cert, ca.ca, pubKey, ca.privateKey)
	if err != nil {
		return exit(codes.Internal, "failed to generate new certificate", err)
	}

	if err := ca.repository.Create(ctx, req.Service, req.PublicKey, signedCert, cert.NotAfter); err != nil {
		return exit(codes.Internal, "failed to write certificate to DB", err,
			slog.String("certificate", string(signedCert)),
		)
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

	start := time.Now()
	defer func() {
		ca.metrics.ObserveServiceCertsFetchLatency(ctx, req.Service, time.Since(start))
	}()

	ca.metrics.IncServiceCertsFetched(req.Service)
	ca.logger.DebugContext(ctx, "new certificate request",
		slog.String("service", req.Service), slog.Bool("with_csr", req.SigningRequest != nil))

	exit := withExit[pb.CertificateRequest, pb.CertificateResponse](
		ctx, ca, req, func() { ca.metrics.IncServiceCertsFetchFailed(req.Service) }, span,
	)

	if err := req.ValidateAll(); err != nil {
		return exit(codes.InvalidArgument, "invalid request", err)
	}

	pubKey, cert, err := ca.repository.Get(ctx, req.Service)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			ca.logger.DebugContext(ctx, "service was not found",
				slog.String("error", err.Error()), slog.String("service", req.Service))

			return nil, status.Error(codes.NotFound, err.Error())
		}

		return exit(codes.Internal, "failed to fetch service from the DB", err)
	}

	if !slices.Equal(pubKey, req.PublicKey) {
		return exit(codes.PermissionDenied, "mismatching public keys", ErrInvalidPublicKey)
	}

	return &pb.CertificateResponse{Certificate: cert}, nil
}

func (ca *CertificateAuthority) VerifyCertificate(ctx context.Context, req *pb.VerificationRequest) (*pb.VerificationResponse, error) {
	ctx, span := ca.tracer.Start(ctx, "CertificateAuthority.VerifyCertificate", trace.WithAttributes(
		attribute.String("service", req.Service),
		attribute.String("certificate", string(req.Certificate)),
	))
	defer span.End()

	start := time.Now()
	defer func() {
		ca.metrics.ObserveVerificationLatency(ctx, req.Service, time.Since(start))
	}()

	ca.metrics.IncVerificationRequests(req.Service)
	ca.logger.DebugContext(ctx, "certificate verification request",
		slog.String("service", req.Service))

	exit := withExit[pb.VerificationRequest, pb.VerificationResponse](
		ctx, ca, req, func() { ca.metrics.IncVerificationFailed(req.Service) }, span,
	)

	if err := req.ValidateAll(); err != nil {
		return exit(codes.InvalidArgument, "invalid request", err)
	}

	pubKey, _, err := ca.repository.Get(ctx, req.Service)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			ca.logger.DebugContext(ctx, "service was not found",
				slog.String("error", err.Error()), slog.String("service", req.Service))

			return nil, status.Error(codes.NotFound, err.Error())
		}

		return exit(codes.Internal, "failed to fetch service from the DB", err)
	}

	storedPub, err := keygen.DecodePublic(pubKey)
	if err != nil {
		return exit(codes.Internal, "failed to decode stored public key", err)
	}

	cert, err := keygen.DecodeCertificate(req.Certificate)
	if err != nil {
		return exit(codes.InvalidArgument, "failed to decode certificate", err)
	}

	pub, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return exit(codes.InvalidArgument, "failed to retrieve public key from certificate", ErrInvalidCertificate)
	}

	if !pub.Equal(storedPub) {
		return exit(codes.PermissionDenied, "mismatching public keys", ErrInvalidPublicKey)
	}

	if time.Now().After(cert.NotAfter) {
		ca.logger.DebugContext(ctx, "expired certificate",
			slog.Time("expiry", cert.NotAfter), slog.String("service", req.Service))

		return &pb.VerificationResponse{Reason: "expired"}, nil
	}

	return &pb.VerificationResponse{Valid: true}, nil
}

func (ca *CertificateAuthority) DeleteService(ctx context.Context, req *pb.DeletionRequest) (*pb.DeletionResponse, error) {
	ctx, span := ca.tracer.Start(ctx, "CertificateAuthority.DeleteService", trace.WithAttributes(
		attribute.String("service", req.Service),
		attribute.String("pub_key", string(req.PublicKey)),
	))
	defer span.End()

	start := time.Now()
	defer func() {
		ca.metrics.ObserveServiceDeletionLatency(ctx, time.Since(start))
	}()

	ca.metrics.IncServiceDeletions()
	ca.logger.DebugContext(ctx, "service deletion request",
		slog.String("service", req.Service))

	exit := withExit[pb.DeletionRequest, pb.DeletionResponse](
		ctx, ca, req, ca.metrics.IncServiceDeletionFailed, span,
	)

	if err := req.ValidateAll(); err != nil {
		return exit(codes.InvalidArgument, "invalid request", err)
	}

	pubKey, _, err := ca.repository.Get(ctx, req.Service)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			ca.logger.DebugContext(ctx, "service was not found",
				slog.String("error", err.Error()), slog.String("service", req.Service))

			return nil, status.Error(codes.NotFound, err.Error())
		}

		return exit(codes.Internal, "failed to fetch service from the DB", err)
	}

	if !slices.Equal(pubKey, req.PublicKey) {
		return exit(codes.PermissionDenied, "mismatching public keys", ErrInvalidPublicKey)
	}

	if err = ca.repository.Delete(ctx, req.Service); err != nil {
		return exit(codes.Internal, "failed to remove service from DB", err)
	}

	return &pb.DeletionResponse{}, nil
}

func (ca *CertificateAuthority) PublicKey(ctx context.Context, _ *pb.PublicKeyRequest) (*pb.PublicKeyResponse, error) {
	ctx, span := ca.tracer.Start(ctx, "CertificateAuthority.PublicKeyRequest")
	defer span.End()

	start := time.Now()
	defer func() {
		ca.metrics.ObservePubKeyRequestLatency(ctx, time.Since(start))
	}()

	ca.metrics.IncPubKeyRequests()
	ca.logger.DebugContext(ctx, "CA's public key request")

	exit := withExit[pb.PublicKeyRequest, pb.PublicKeyResponse](
		ctx, ca, nil, ca.metrics.IncPubKeyRequestFailed, span,
	)

	key, err := keygen.EncodePublic(&ca.privateKey.PublicKey)
	if err != nil {
		return exit(codes.Internal, "failed to encode CA's public key", err)
	}

	return &pb.PublicKeyResponse{PublicKey: key}, nil
}

func withExit[Req any, Res any](
	ctx context.Context, ca *CertificateAuthority,
	req *Req, metric func(), span trace.Span,
) func(codes.Code, string, error, ...any) (*Res, error) {
	return func(code codes.Code, message string, err error, args ...any) (*Res, error) {
		logArgs := make([]any, 0, len(args)+2)

		if req != nil {
			logArgs = append(logArgs, slog.Any("request", req))
		}

		if err != nil {
			logArgs = append(logArgs, slog.String("error", err.Error()))
			logArgs = append(logArgs, args...)

			span.SetStatus(otelcodes.Error, err.Error())
			span.RecordError(err)
			metric()
			ca.logger.WarnContext(ctx, message, logArgs...)

			return nil, status.Error(code, err.Error())
		}

		logArgs = append(logArgs, args...)
		span.SetStatus(otelcodes.Error, message)
		span.RecordError(errors.New(message))
		metric()
		ca.logger.WarnContext(ctx, message, logArgs...)

		return nil, status.Error(code, message)
	}
}
