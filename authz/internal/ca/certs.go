package ca

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zalgonoise/x/reg"

	"github.com/zalgonoise/x/authz/internal/certs"
	"github.com/zalgonoise/x/authz/internal/keygen"
	pb "github.com/zalgonoise/x/authz/pb/authz/v1"
)

func (ca *CertificateAuthority) CreateCertificate(ctx context.Context, req *pb.CertificateRequest) (*pb.CertificateResponse, error) {
	ctx, span := ca.tracer.Start(ctx, "CertificateAuthority.CreateCertificate", trace.WithAttributes(
		attribute.String("service", req.Service),
		attribute.String("pub_key", string(req.PublicKey)),
		attribute.Bool("with_csr", req.SigningRequest != nil),
	))
	defer span.End()

	start := time.Now()
	defer func() {
		ca.metrics.ObserveCertificatesCreateLatency(ctx, req.Service, time.Since(start))
	}()

	ca.metrics.IncCertificatesCreated(req.Service)
	ca.logger.DebugContext(ctx, "new certificate creation request",
		slog.String("service", req.Service), slog.Bool("with_csr", req.SigningRequest != nil))

	if err := req.ValidateAll(); err != nil {
		ca.r.Event(ctx, "invalid request",
			reg.WithError(err),
			reg.WithLogLevel(slog.LevelWarn),
			reg.WithLogAttributes(slog.Any("request", req)),
			reg.WithSpan(span),
			reg.WithMetric(func() { ca.metrics.IncCertificatesCreateFailed(req.Service) }),
		)

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	certificates, err := ca.repository.ListCertificates(ctx, req.Service)
	if err != nil {
		ca.r.Event(ctx, "failed to get service certificates from DB",
			reg.WithError(err),
			reg.WithLogAttributes(slog.String("service", req.Service)),
			reg.WithSpan(span),
			reg.WithMetric(func() {
				ca.metrics.IncServiceRegistryFailed()
				ca.metrics.IncCertificatesCreateFailed(req.Service)
			}),
		)

		return nil, status.Error(codes.Internal, err.Error())
	}

	// max number of certificates, return the stored certificate with the biggest validity
	if len(certificates) >= certificateLimit {
		return certificates[0], nil
	}

	certificate, expiry, err := ca.newCertificate(ctx, req)
	if err != nil {
		ca.r.Event(ctx, "creating certificate",
			reg.WithError(err),
			reg.WithLogAttributes(slog.String("service", req.Service)),
			reg.WithSpan(span),
			reg.WithMetric(func() {
				ca.metrics.IncServiceRegistryFailed()
				ca.metrics.IncCertificatesCreateFailed(req.Service)
			}),
		)

		if errors.Is(err, ErrInvalidPublicKey) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := ca.repository.CreateCertificate(ctx, req.Service, certificate, expiry); err != nil {
		ca.r.Event(ctx, "writing certificate to database",
			reg.WithError(err),
			reg.WithLogAttributes(slog.String("service", req.Service)),
			reg.WithSpan(span),
			reg.WithMetric(func() {
				ca.metrics.IncServiceRegistryFailed()
				ca.metrics.IncCertificatesCreateFailed(req.Service)
			}),
		)

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CertificateResponse{
		Certificate: certificate,
		ExpiresOn:   expiry.UnixMilli(),
	}, nil
}

func (ca *CertificateAuthority) ListCertificates(ctx context.Context, req *pb.CertificateRequest) (*pb.ListCertificatesResponse, error) {
	ctx, span := ca.tracer.Start(ctx, "CertificateAuthority.ListCertificates", trace.WithAttributes(
		attribute.String("service", req.Service),
		attribute.String("pub_key", string(req.PublicKey)),
		attribute.Bool("with_csr", req.SigningRequest != nil),
	))
	defer span.End()

	start := time.Now()
	defer func() {
		ca.metrics.ObserveCertificatesListLatency(ctx, req.Service, time.Since(start))
	}()

	ca.metrics.IncCertificatesListed(req.Service)
	ca.logger.DebugContext(ctx, "new certificate listing request",
		slog.String("service", req.Service), slog.Bool("with_csr", req.SigningRequest != nil))

	if err := req.ValidateAll(); err != nil {
		ca.r.Event(ctx, "invalid request",
			reg.WithError(err),
			reg.WithLogLevel(slog.LevelWarn),
			reg.WithLogAttributes(slog.String("service", req.Service)),
			reg.WithSpan(span),
			reg.WithMetric(func() { ca.metrics.IncCertificatesCreateFailed(req.Service) }),
		)

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err := ca.validatePublicKeys(ctx, req.Service, req.PublicKey)
	switch {
	case errors.Is(err, ErrInvalidPublicKey):
		ca.r.Event(ctx, "mismatching public keys",
			reg.WithError(err),
			reg.WithLogLevel(slog.LevelWarn),
			reg.WithLogAttributes(slog.String("service", req.Service)),
			reg.WithSpan(span),
			reg.WithMetric(func() { ca.metrics.IncCertificatesListFailed(req.Service) }),
		)

		return nil, status.Error(codes.PermissionDenied, err.Error())
	case err != nil:
		ca.r.Event(ctx, "validating public keys",
			reg.WithError(err),
			reg.WithLogAttributes(slog.String("service", req.Service)),
			reg.WithSpan(span),
			reg.WithMetric(func() { ca.metrics.IncCertificatesListFailed(req.Service) }),
		)

		return nil, status.Error(codes.Internal, err.Error())
	default:
	}

	certificates, err := ca.repository.ListCertificates(ctx, req.Service)
	if err != nil {
		ca.r.Event(ctx, "getting service certificates from database",
			reg.WithError(err),
			reg.WithLogAttributes(slog.String("service", req.Service)),
			reg.WithSpan(span),
			reg.WithMetric(func() { ca.metrics.IncCertificatesListFailed(req.Service) }),
		)

		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(certificates) == 0 {
		ca.r.Event(ctx, "getting service certificates from database",
			reg.WithError(err),
			reg.WithLogLevel(slog.LevelWarn),
			reg.WithLogAttributes(slog.String("service", req.Service)),
			reg.WithSpan(span),
			reg.WithMetric(func() { ca.metrics.IncCertificatesListFailed(req.Service) }),
		)

		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &pb.ListCertificatesResponse{Certificates: certificates}, nil
}

func (ca *CertificateAuthority) DeleteCertificate(ctx context.Context, req *pb.CertificateDeletionRequest) (*pb.CertificateDeletionResponse, error) {
	ctx, span := ca.tracer.Start(ctx, "CertificateAuthority.DeleteCertificate", trace.WithAttributes(
		attribute.String("service", req.Service),
		attribute.String("certificate", string(req.Certificate)),
	))
	defer span.End()

	start := time.Now()
	defer func() {
		ca.metrics.ObserveCertificatesDeleteLatency(ctx, req.Service, time.Since(start))
	}()

	ca.metrics.IncCertificatesDeleted(req.Service)
	ca.logger.DebugContext(ctx, "certificate deletion request",
		slog.String("service", req.Service))

	if err := req.ValidateAll(); err != nil {
		ca.r.Event(ctx, "invalid request",
			reg.WithError(err),
			reg.WithLogLevel(slog.LevelWarn),
			reg.WithLogAttributes(slog.Any("request", req)),
			reg.WithSpan(span),
			reg.WithMetric(func() { ca.metrics.IncCertificatesDeleteFailed(req.Service) }),
		)

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := ca.validatePublicKeys(ctx, req.Service, req.PublicKey); err != nil {
		ca.r.Event(ctx, "validating public keys",
			reg.WithError(err),
			reg.WithLogLevel(slog.LevelWarn),
			reg.WithLogAttributes(slog.String("service", req.Service)),
			reg.WithSpan(span),
			reg.WithMetric(func() { ca.metrics.IncCertificatesDeleteFailed(req.Service) }),
		)

		return nil, status.Error(codes.InvalidArgument, ErrInvalidPublicKey.Error())
	}

	if err := certs.Verify(req.Certificate, ca.ca, nil); err != nil {
		ca.r.Event(ctx, "validating certificate",
			reg.WithError(err),
			reg.WithLogLevel(slog.LevelWarn),
			reg.WithLogAttributes(slog.String("service", req.Service)),
			reg.WithSpan(span),
			reg.WithMetric(func() { ca.metrics.IncCertificatesDeleteFailed(req.Service) }),
		)

		return nil, status.Error(codes.InvalidArgument, ErrInvalidCertificate.Error())
	}

	if err := ca.repository.DeleteCertificate(ctx, req.Service, req.Certificate); err != nil {
		ca.r.Event(ctx, "removing stored certificate",
			reg.WithError(err),
			reg.WithLogAttributes(slog.String("service", req.Service)),
			reg.WithSpan(span),
			reg.WithMetric(func() { ca.metrics.IncCertificatesDeleteFailed(req.Service) }),
		)

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CertificateDeletionResponse{}, nil
}

func (ca *CertificateAuthority) VerifyCertificate(ctx context.Context, req *pb.VerificationRequest) (*pb.VerificationResponse, error) {
	ctx, span := ca.tracer.Start(ctx, "CertificateAuthority.VerifyCertificate", trace.WithAttributes(
		attribute.String("service", req.Service),
		attribute.String("certificate", string(req.Certificate)),
	))
	defer span.End()

	start := time.Now()
	defer func() {
		ca.metrics.ObserveCertificateVerificationLatency(ctx, req.Service, time.Since(start))
	}()

	ca.metrics.IncCertificatesDeleteFailed(req.Service)
	ca.logger.DebugContext(ctx, "certificate verification request",
		slog.String("service", req.Service))

	if err := req.ValidateAll(); err != nil {
		ca.r.Event(ctx, "invalid request",
			reg.WithError(err),
			reg.WithLogLevel(slog.LevelWarn),
			reg.WithLogAttributes(slog.Any("request", req)),
			reg.WithSpan(span),
			reg.WithMetric(func() { ca.metrics.IncCertificateVerificationFailed(req.Service) }),
		)

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := certs.Verify(req.Certificate, ca.ca, nil); err != nil {
		ca.r.Event(ctx, "verifying certificate",
			reg.WithError(err),
			reg.WithLogLevel(slog.LevelWarn),
			reg.WithLogAttributes(slog.String("service", req.Service)),
			reg.WithSpan(span),
			reg.WithMetric(func() { ca.metrics.IncCertificateVerificationFailed(req.Service) }),
		)

		return nil, status.Error(codes.InvalidArgument, ErrInvalidCertificate.Error())
	}

	return &pb.VerificationResponse{}, nil
}

func (ca *CertificateAuthority) RootCertificate(ctx context.Context, _ *pb.RootCertificateRequest) (*pb.RootCertificateResponse, error) {
	ctx, span := ca.tracer.Start(ctx, "CertificateAuthority.RootCertificate")
	defer span.End()

	start := time.Now()
	defer func() {
		ca.metrics.ObserveRootCertificateRequestLatency(ctx, time.Since(start))
	}()

	ca.metrics.IncRootCertificateRequests()
	ca.logger.DebugContext(ctx, "CA certificate request")

	return &pb.RootCertificateResponse{Root: ca.raw}, nil
}

func (ca *CertificateAuthority) newCertificate(ctx context.Context, req *pb.CertificateRequest) ([]byte, time.Time, error) {
	pubKey, err := keygen.DecodePublic(req.PublicKey)
	if err != nil {
		ca.logger.WarnContext(ctx, "decoding public key",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, time.Time{}, ErrInvalidPublicKey
	}

	cert, err := certs.NewCertFromCSR(ca.ca.Version, ca.durMonth, ca.ca.Subject, certs.ToCSR(req.Service, pubKey, req.SigningRequest))
	if err != nil {
		ca.logger.ErrorContext(ctx, "generating new certificate",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, time.Time{}, err
	}

	signedCert, err := certs.Encode(cert, ca.ca, pubKey, ca.privateKey)
	if err != nil {
		ca.logger.ErrorContext(ctx, "encoding the new certificate",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, time.Time{}, err
	}

	return signedCert, cert.NotAfter, nil
}
