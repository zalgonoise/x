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

	exit := withExit[pb.CertificateRequest, pb.CertificateResponse](
		ctx, a.logger, req, func() { a.metrics.IncServiceCertsFetchFailed(req.Service) }, span,
	)

	if err := req.ValidateAll(); err != nil {
		return exit(codes.InvalidArgument, "invalid request", err)
	}

	pubKey, cert, err := a.services.GetService(ctx, req.Service)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			a.logger.DebugContext(ctx, "service was not found",
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

	exit := withExit[pb.VerificationRequest, pb.VerificationResponse](
		ctx, a.logger, req, func() { a.metrics.IncVerificationFailed(req.Service) }, span,
	)

	if err := req.ValidateAll(); err != nil {
		return exit(codes.InvalidArgument, "invalid request", err)
	}

	pubKey, _, err := a.services.GetService(ctx, req.Service)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			a.logger.DebugContext(ctx, "service was not found",
				slog.String("error", err.Error()), slog.String("service", req.Service))

			return nil, status.Error(codes.NotFound, err.Error())
		}

		return exit(codes.Internal, "failed to fetch service from the DB", err)
	}

	storedPub, err := keygen.DecodePublic(pubKey)
	if err != nil {
		return exit(codes.Internal, "failed to decode stored public key", err)
	}

	cert, err := certs.Decode(req.Certificate)
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

	exit := withExit[pb.DeletionRequest, pb.DeletionResponse](
		ctx, a.logger, req, a.metrics.IncServiceDeletionFailed, span,
	)

	if err := req.ValidateAll(); err != nil {
		return exit(codes.InvalidArgument, "invalid request", err)
	}

	pubKey, _, err := a.services.GetService(ctx, req.Service)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			a.logger.DebugContext(ctx, "service was not found",
				slog.String("error", err.Error()), slog.String("service", req.Service))

			return nil, status.Error(codes.NotFound, err.Error())
		}

		return exit(codes.Internal, "failed to fetch service from the DB", err)
	}

	if !slices.Equal(pubKey, req.PublicKey) {
		return exit(codes.PermissionDenied, "mismatching public keys", ErrInvalidPublicKey)
	}

	if err = a.services.DeleteService(ctx, req.Service); err != nil {
		return exit(codes.Internal, "failed to remove service from DB", err)
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

	exit := withExit[pb.PublicKeyRequest, pb.PublicKeyResponse](
		ctx, a.logger, nil, a.metrics.IncPubKeyRequestFailed, span,
	)

	key, err := keygen.EncodePublic(&a.privateKey.PublicKey)
	if err != nil {
		return exit(codes.Internal, "failed to encode authz service's public key", err)
	}

	return &pb.PublicKeyResponse{PublicKey: key}, nil
}
