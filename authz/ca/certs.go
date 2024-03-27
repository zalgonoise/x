package ca

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/zalgonoise/x/authz/certs"
	"github.com/zalgonoise/x/authz/keygen"
	pb "github.com/zalgonoise/x/authz/pb/authz/v1"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		ca.metrics.IncCertificatesCreateFailed(req.Service)

		ca.logger.WarnContext(ctx, "invalid request",
			slog.Any("request", req), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	certificates, err := ca.repository.ListCertificates(ctx, req.Service)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		ca.metrics.IncServiceRegistryFailed()
		ca.metrics.IncCertificatesCreateFailed(req.Service)

		ca.logger.ErrorContext(ctx, "failed to get service certificates from DB",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	// max number of certificates, return the stored certificate with the biggest validity
	if len(certificates) >= certificateLimit {
		return certificates[0], nil
	}

	certificate, expiry, err := ca.newCertificate(ctx, req)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		ca.metrics.IncServiceRegistryFailed()
		ca.metrics.IncCertificatesCreateFailed(req.Service)

		if errors.Is(err, ErrInvalidPublicKey) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := ca.repository.CreateCertificate(ctx, req.Service, certificate, expiry); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		ca.metrics.IncServiceRegistryFailed()
		ca.metrics.IncCertificatesCreateFailed(req.Service)

		ca.logger.ErrorContext(ctx, "failed to write certificate to DB",
			slog.String("service", req.Service), slog.String("error", err.Error()))

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
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		ca.metrics.IncCertificatesListFailed(req.Service)

		ca.logger.WarnContext(ctx, "invalid request",
			slog.Any("request", req), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := ca.validatePublicKeys(ctx, req.Service, req.PublicKey); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		ca.metrics.IncCertificatesListFailed(req.Service)

		if errors.Is(err, ErrInvalidPublicKey) {
			ca.logger.WarnContext(ctx, "mismatching public keys",
				slog.String("service", req.Service), slog.String("error", err.Error()))

			return nil, status.Error(codes.PermissionDenied, err.Error())
		}

		ca.logger.ErrorContext(ctx, "failed to validate public keys",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	certificates, err := ca.repository.ListCertificates(ctx, req.Service)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		ca.metrics.IncCertificatesListFailed(req.Service)

		ca.logger.ErrorContext(ctx, "failed to get service certificates from DB",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(certificates) == 0 {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		ca.metrics.IncCertificatesListFailed(req.Service)

		ca.logger.WarnContext(ctx, "no certificates were found",
			slog.String("service", req.Service), slog.String("error", err.Error()))

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
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		ca.metrics.IncCertificatesDeleteFailed(req.Service)

		ca.logger.WarnContext(ctx, "invalid request",
			slog.Any("request", req), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := ca.validatePublicKeys(ctx, req.Service, req.PublicKey); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		ca.metrics.IncCertificatesDeleteFailed(req.Service)

		ca.logger.WarnContext(ctx, "invalid public keys",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, ErrInvalidPublicKey.Error())
	}

	if err := certs.Verify(req.Certificate, ca.ca, nil); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		ca.metrics.IncCertificatesDeleteFailed(req.Service)

		ca.logger.WarnContext(ctx, "invalid certificate",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, ErrInvalidCertificate.Error())
	}

	if err := ca.repository.DeleteCertificate(ctx, req.Service, req.Certificate); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		ca.metrics.IncCertificatesDeleteFailed(req.Service)

		ca.logger.WarnContext(ctx, "failed to remove stored certificate",
			slog.String("service", req.Service), slog.String("error", err.Error()))

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
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		ca.metrics.IncCertificateVerificationFailed(req.Service)

		ca.logger.WarnContext(ctx, "invalid request",
			slog.Any("request", req), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := certs.Verify(req.Certificate, ca.ca, nil); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		ca.metrics.IncCertificateVerificationFailed(req.Service)

		ca.logger.WarnContext(ctx, "invalid certificate",
			slog.String("service", req.Service), slog.String("error", err.Error()))

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
		ca.logger.WarnContext(ctx, "invalid public key",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, time.Time{}, ErrInvalidPublicKey
	}

	cert, err := certs.NewCertFromCSR(ca.ca.Version, ca.durMonth, ca.ca.Subject, certs.ToCSR(req.Service, pubKey, req.SigningRequest))
	if err != nil {
		ca.logger.ErrorContext(ctx, "failed to generate new certificate",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, time.Time{}, err
	}

	signedCert, err := certs.Encode(cert, ca.ca, pubKey, ca.privateKey)
	if err != nil {
		ca.logger.ErrorContext(ctx, "failed to encode the new certificate",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, time.Time{}, err
	}

	return signedCert, cert.NotAfter, nil
}
