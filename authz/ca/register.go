package ca

import (
	"context"
	"errors"
	"log/slog"
	"time"

	pb "github.com/zalgonoise/x/authz/pb/authz/v1"
	"github.com/zalgonoise/x/authz/repository"
	"github.com/zalgonoise/x/reg"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (ca *CertificateAuthority) RegisterService(
	ctx context.Context, req *pb.CertificateRequest) (*pb.CertificateResponse, error) {
	ctx, span := ca.tracer.Start(ctx, "CertificateAuthority.RegisterService", trace.WithAttributes(
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

	if err := req.ValidateAll(); err != nil {
		ca.r.Event(ctx, "invalid request",
			reg.WithError(err),
			reg.WithLogLevel(slog.LevelWarn),
			reg.WithLogAttributes(slog.Any("request", req)),
			reg.WithSpan(span),
			reg.WithMetric(ca.metrics.IncServiceRegistryFailed),
		)

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	switch err := ca.validatePublicKeys(ctx, req.Service, req.PublicKey); {
	// create the service if not exists
	case errors.Is(err, repository.ErrNotFound):
		if err := ca.repository.CreateService(ctx, req.Service, req.PublicKey); err != nil {
			ca.r.Event(ctx, "failed to write service to DB",
				reg.WithError(err),
				reg.WithLogAttributes(slog.String("service", req.Service)),
				reg.WithSpan(span),
				reg.WithMetric(ca.metrics.IncServiceRegistryFailed),
			)

			return nil, status.Error(codes.Internal, err.Error())
		}

	// service exists, invalid public keys
	case errors.Is(err, ErrInvalidPublicKey):
		ca.r.Event(ctx, "mismatching public keys",
			reg.WithError(err),
			reg.WithLogLevel(slog.LevelWarn),
			reg.WithLogAttributes(slog.String("service", req.Service)),
			reg.WithSpan(span),
			reg.WithMetric(ca.metrics.IncServiceRegistryFailed),
		)

		return nil, status.Error(codes.PermissionDenied, err.Error())

	// internal error
	case err != nil:
		ca.r.Event(ctx, "failed to validate public keys",
			reg.WithError(err),
			reg.WithLogAttributes(slog.String("service", req.Service)),
			reg.WithSpan(span),
			reg.WithMetric(ca.metrics.IncServiceRegistryFailed),
		)

		return nil, status.Error(codes.Internal, err.Error())
	default:
	}

	return ca.CreateCertificate(ctx, req)
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

	if err := req.ValidateAll(); err != nil {
		ca.r.Event(ctx, "invalid request",
			reg.WithError(err),
			reg.WithLogLevel(slog.LevelWarn),
			reg.WithLogAttributes(slog.Any("request", req)),
			reg.WithSpan(span),
			reg.WithMetric(ca.metrics.IncServiceDeletionFailed),
		)

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := ca.validatePublicKeys(ctx, req.Service, req.PublicKey); err != nil {
		ca.r.Event(ctx, "mismatching public keys",
			reg.WithError(err),
			reg.WithLogLevel(slog.LevelWarn),
			reg.WithLogAttributes(slog.String("service", req.Service)),
			reg.WithSpan(span),
			reg.WithMetric(ca.metrics.IncServiceDeletionFailed),
		)

		return nil, status.Error(codes.PermissionDenied, ErrInvalidPublicKey.Error())
	}

	if err := ca.repository.DeleteService(ctx, req.Service); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		ca.metrics.IncServiceDeletionFailed()

		ca.logger.ErrorContext(ctx, "failed to remove service from DB",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeletionResponse{}, nil
}
