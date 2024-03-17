package authz

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"log/slog"
	"slices"
	"time"

	"github.com/zalgonoise/x/authz/certs"
	"github.com/zalgonoise/x/authz/keygen"
	pb "github.com/zalgonoise/x/authz/pb/authz/v1"
	"github.com/zalgonoise/x/authz/repository"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a *Authz) GetCertificate(ctx context.Context, req *pb.CertificateRequest) (*pb.CertificateResponse, error) {
	ctx, span := a.tracer.Start(ctx, "Authz.GetCertificate", trace.WithAttributes(
		attribute.String("service", req.Service),
		attribute.String("pub_key", string(req.PublicKey)),
		attribute.Bool("with_csr", req.SigningRequest != nil),
	))
	defer span.End()

	start := time.Now()
	defer func() {
		a.metrics.ObserveServiceCertsFetchLatency(ctx, req.Service, time.Since(start))
	}()

	a.metrics.IncServiceCertsFetched(req.Service)
	a.logger.DebugContext(ctx, "new certificate request",
		slog.String("service", req.Service), slog.Bool("with_csr", req.SigningRequest != nil))

	if err := req.ValidateAll(); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceCertsFetchFailed("")

		a.logger.WarnContext(ctx, "invalid request",
			slog.Any("request", req), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	pubKey, cert, err := a.services.GetService(ctx, req.Service)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceCertsFetchFailed(req.Service)

		if errors.Is(err, repository.ErrNotFound) {
			a.logger.DebugContext(ctx, "service was not found",
				slog.String("service", req.Service), slog.String("error", err.Error()))

			return nil, status.Error(codes.NotFound, err.Error())
		}

		a.logger.ErrorContext(ctx, "failed to fetch service from the DB",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	if !slices.Equal(pubKey, req.PublicKey) {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceCertsFetchFailed(req.Service)

		a.logger.WarnContext(ctx, "mismatching public keys",
			slog.Any("request", req), slog.String("error", err.Error()))

		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	return &pb.CertificateResponse{Certificate: cert}, nil
}

func (a *Authz) VerifyCertificate(ctx context.Context, req *pb.VerificationRequest) (*pb.VerificationResponse, error) {
	ctx, span := a.tracer.Start(ctx, "Authz.VerifyCertificate", trace.WithAttributes(
		attribute.String("service", req.Service),
		attribute.String("certificate", string(req.Certificate)),
	))
	defer span.End()

	start := time.Now()
	defer func() {
		a.metrics.ObserveVerificationLatency(ctx, req.Service, time.Since(start))
	}()

	a.metrics.IncVerificationRequests(req.Service)
	a.logger.DebugContext(ctx, "certificate verification request",
		slog.String("service", req.Service))

	if err := req.ValidateAll(); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncVerificationFailed("")

		a.logger.WarnContext(ctx, "invalid request",
			slog.Any("request", req), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	pubKey, _, err := a.services.GetService(ctx, req.Service)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncVerificationFailed(req.Service)

		if errors.Is(err, repository.ErrNotFound) {
			a.logger.DebugContext(ctx, "service was not found",
				slog.String("service", req.Service), slog.String("error", err.Error()))

			return nil, status.Error(codes.NotFound, ErrInvalidService.Error())
		}

		a.logger.ErrorContext(ctx, "failed to fetch service from the DB",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	storedPub, err := keygen.DecodePublic(pubKey)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncVerificationFailed(req.Service)

		a.logger.ErrorContext(ctx, "failed to decode stored public key",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	cert, err := certs.Decode(req.Certificate)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncVerificationFailed(req.Service)

		a.logger.WarnContext(ctx, "failed to decode certificate",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, ErrInvalidIDCertificate.Error())
	}

	pub, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncVerificationFailed(req.Service)

		a.logger.WarnContext(ctx, "failed to retrieve public key from certificate",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, ErrInvalidIDCertificate.Error())
	}

	if !pub.Equal(storedPub) {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncVerificationFailed(req.Service)

		a.logger.WarnContext(ctx, "mismatching public keys",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, status.Error(codes.PermissionDenied, ErrInvalidPublicKey.Error())
	}

	if time.Now().After(cert.NotAfter) {
		a.logger.DebugContext(ctx, "expired certificate",
			slog.Time("expiry", cert.NotAfter), slog.String("service", req.Service))

		return &pb.VerificationResponse{Reason: "expired"}, nil
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

	pubKey, _, err := a.services.GetService(ctx, req.Service)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			a.logger.DebugContext(ctx, "service was not found",
				slog.String("service", req.Service), slog.String("error", err.Error()))

			return &pb.DeletionResponse{}, nil
		}

		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceDeletionFailed()

		a.logger.ErrorContext(ctx, "failed to fetch service from the DB",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	if !slices.Equal(pubKey, req.PublicKey) {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceDeletionFailed()

		a.logger.ErrorContext(ctx, "mismatching public keys",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, status.Error(codes.PermissionDenied, ErrInvalidPublicKey.Error())
	}

	if err = a.services.DeleteService(ctx, req.Service); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceDeletionFailed()

		a.logger.ErrorContext(ctx, "failed to remove service from DB",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeletionResponse{}, nil
}

func (a *Authz) PublicKey(ctx context.Context, req *pb.PublicKeyRequest) (*pb.PublicKeyResponse, error) {
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
