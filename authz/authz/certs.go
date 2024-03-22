package authz

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

func (a *Authz) ListCertificates(ctx context.Context, req *pb.CertificateRequest) (*pb.ListCertificatesResponse, error) {
	ctx, span := a.tracer.Start(ctx, "Authz.ListCertificates", trace.WithAttributes(
		attribute.String("service", req.Service),
		attribute.String("pub_key", string(req.PublicKey)),
		attribute.Bool("with_csr", req.SigningRequest != nil),
	))
	defer span.End()

	start := time.Now()
	defer func() {
		a.metrics.ObserveCertificatesListLatency(ctx, req.Service, time.Since(start))
	}()

	a.metrics.IncCertificatesListed(req.Service)
	a.logger.DebugContext(ctx, "new certificate listing request",
		slog.String("service", req.Service), slog.Bool("with_csr", req.SigningRequest != nil))

	if err := req.ValidateAll(); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncCertificatesListFailed("")

		a.logger.WarnContext(ctx, "invalid request",
			slog.Any("request", req), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := a.validatePublicKeys(ctx, req.Service, req.PublicKey); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncCertificatesListFailed(req.Service)

		if errors.Is(err, ErrInvalidPublicKey) {
			a.logger.WarnContext(ctx, "mismatching public keys",
				slog.String("service", req.Service), slog.String("error", err.Error()))

			return nil, status.Error(codes.PermissionDenied, err.Error())
		}

		a.logger.ErrorContext(ctx, "failed to validate public keys",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	certificates, err := a.services.ListCertificates(ctx, req.Service)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncCertificatesListFailed(req.Service)

		a.logger.ErrorContext(ctx, "failed to get service certificates from DB",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(certificates) == 0 {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncCertificatesListFailed(req.Service)

		a.logger.WarnContext(ctx, "no certificates were found",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &pb.ListCertificatesResponse{Certificates: certificates}, nil
}

func (a *Authz) DeleteCertificate(ctx context.Context, req *pb.CertificateDeletionRequest) (*pb.CertificateDeletionResponse, error) {
	ctx, span := a.tracer.Start(ctx, "Authz.DeleteCertificate", trace.WithAttributes(
		attribute.String("service", req.Service),
		attribute.String("certificate", string(req.Certificate)),
	))
	defer span.End()

	start := time.Now()
	defer func() {
		a.metrics.ObserveCertificatesDeleteLatency(ctx, req.Service, time.Since(start))
	}()

	a.metrics.IncCertificatesDeleted(req.Service)
	a.logger.DebugContext(ctx, "certificate deletion request",
		slog.String("service", req.Service))

	if err := req.ValidateAll(); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncCertificatesDeleteFailed(req.Service)

		a.logger.WarnContext(ctx, "invalid request",
			slog.Any("request", req), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := a.validatePublicKeys(ctx, req.Service, req.PublicKey); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncCertificatesDeleteFailed(req.Service)

		a.logger.WarnContext(ctx, "invalid public keys",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, ErrInvalidPublicKey.Error())
	}

	caCert, err := certs.Decode(a.caCert)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncCertificateVerificationFailed(req.Service)

		a.logger.WarnContext(ctx, "failed to decode CA certificate",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := certs.Verify(req.Certificate, a.cert, caCert); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncCertificatesDeleteFailed(req.Service)

		a.logger.WarnContext(ctx, "invalid certificate",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, ErrInvalidIDCertificate.Error())
	}

	if err := a.services.DeleteCertificate(ctx, req.Service, req.Certificate); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncCertificatesDeleteFailed(req.Service)

		a.logger.WarnContext(ctx, "failed to remove stored certificate",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CertificateDeletionResponse{}, nil
}

func (a *Authz) VerifyCertificate(ctx context.Context, req *pb.VerificationRequest) (*pb.VerificationResponse, error) {
	ctx, span := a.tracer.Start(ctx, "Authz.VerifyCertificate", trace.WithAttributes(
		attribute.String("service", req.Service),
		attribute.String("certificate", string(req.Certificate)),
	))
	defer span.End()

	start := time.Now()
	defer func() {
		a.metrics.ObserveCertificateVerificationLatency(ctx, req.Service, time.Since(start))
	}()

	a.metrics.IncCertificatesVerified(req.Service)
	a.logger.DebugContext(ctx, "certificate verification request",
		slog.String("service", req.Service))

	if err := req.ValidateAll(); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncCertificateVerificationFailed("")

		a.logger.WarnContext(ctx, "invalid request",
			slog.Any("request", req), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	caCert, err := certs.Decode(a.caCert)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncCertificateVerificationFailed(req.Service)

		a.logger.WarnContext(ctx, "failed to decode CA certificate",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := certs.Verify(req.Certificate, a.cert, caCert); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncCertificateVerificationFailed(req.Service)

		a.logger.WarnContext(ctx, "invalid certificate",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, ErrInvalidIDCertificate.Error())
	}

	return &pb.VerificationResponse{Valid: true}, nil
}

func (a *Authz) DeleteService(ctx context.Context, req *pb.DeletionRequest) (*pb.DeletionResponse, error) {
	ctx, span := a.tracer.Start(ctx, "Authz.DeleteService", trace.WithAttributes(
		attribute.String("service", req.Service),
		attribute.String("pub_key", string(req.PublicKey)),
	))
	defer span.End()

	start := time.Now()
	defer func() {
		a.metrics.ObserveServiceDeletionLatency(ctx, time.Since(start))
	}()

	a.metrics.IncServiceDeletions()
	a.logger.DebugContext(ctx, "service deletion request",
		slog.String("service", req.Service))

	if err := req.ValidateAll(); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceDeletionFailed()

		a.logger.ErrorContext(ctx, "invalid request",
			slog.Any("request", req), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := a.validatePublicKeys(ctx, req.Service, req.PublicKey); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceDeletionFailed()

		a.logger.WarnContext(ctx, "mismatching public keys",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, status.Error(codes.PermissionDenied, ErrInvalidPublicKey.Error())
	}

	if err := a.services.DeleteService(ctx, req.Service); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceDeletionFailed()

		a.logger.ErrorContext(ctx, "failed to remove service from DB",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeletionResponse{}, nil
}

func (a *Authz) PublicKey(ctx context.Context, _ *pb.PublicKeyRequest) (*pb.PublicKeyResponse, error) {
	ctx, span := a.tracer.Start(ctx, "Authz.PublicKeyRequest")
	defer span.End()

	start := time.Now()
	defer func() {
		a.metrics.ObservePubKeyRequestLatency(ctx, time.Since(start))
	}()

	a.metrics.IncPubKeyRequests()
	a.logger.DebugContext(ctx, "authz service's public key request")

	key, err := keygen.EncodePublic(&a.privateKey.PublicKey)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncPubKeyRequestFailed()

		a.logger.ErrorContext(ctx, "failed to encode authz service's public key",
			slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.PublicKeyResponse{PublicKey: key}, nil
}
