package authz

import (
	"context"
	"errors"
	"log/slog"
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

// TODO: SignUp / Register must support multiple certificates (up to two) to allow rotation
func (a *Authz) SignUp(ctx context.Context, req *pb.SignUpRequest) (*pb.SignUpResponse, error) {
	ctx, span := a.tracer.Start(ctx, "Authz.SignUp", trace.WithAttributes(
		attribute.String("service", req.Name),
		attribute.String("pub_key", string(req.PublicKey)),
		attribute.Bool("with_csr", req.SigningRequest != nil),
	))
	defer span.End()

	start := time.Now()
	defer func() {
		a.metrics.ObserveServiceRegistryLatency(ctx, time.Since(start))
	}()

	a.metrics.IncServiceRegistries()
	a.logger.DebugContext(ctx, "new sign-up request",
		slog.String("service", req.Name), slog.Bool("with_csr", req.SigningRequest != nil))

	if err := req.ValidateAll(); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceRegistryFailed()

		a.logger.WarnContext(ctx, "invalid request",
			slog.Any("request", req), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	storedPubKey, storedCert, err := a.services.GetService(ctx, req.Name)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceRegistryFailed()

		a.logger.ErrorContext(ctx, "failed to get service from DB",
			slog.String("service", req.Name), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	// service is already registered, just return the stored certificate
	if err == nil && len(storedPubKey) > 0 && len(storedCert) > 0 {
		authzPub, err := keygen.EncodePublic(&a.privateKey.PublicKey)
		if err != nil {
			span.SetStatus(otelcodes.Error, err.Error())
			span.RecordError(err)
			a.metrics.IncServiceRegistryFailed()

			a.logger.ErrorContext(ctx, "failed to encode public key",
				slog.String("service", req.Name), slog.String("error", err.Error()))

			return nil, status.Error(codes.Internal, err.Error())
		}

		return &pb.SignUpResponse{
			Certificate: storedCert,
			Service: &pb.ID{
				PublicKey:   authzPub,
				Certificate: a.caCert,
			},
		}, nil
	}

	pubKey, err := keygen.DecodePublic(req.PublicKey)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceRegistryFailed()

		a.logger.WarnContext(ctx, "invalid public key",
			slog.String("service", req.Name), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	cert, err := certs.NewCertFromCSR(a.cert.Version, a.durMonth, certs.ToCSR(req.Name, pubKey, req.SigningRequest))
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceRegistryFailed()

		a.logger.ErrorContext(ctx, "failed to generate new certificate",
			slog.String("service", req.Name), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	signedCert, err := certs.Encode(cert, a.cert, pubKey, a.privateKey)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceRegistryFailed()

		a.logger.ErrorContext(ctx, "failed to encode the new certificate",
			slog.String("service", req.Name), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := a.services.CreateService(ctx, req.Name, req.PublicKey, signedCert, cert.NotAfter); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceRegistryFailed()

		a.logger.ErrorContext(ctx, "failed to write certificate to DB",
			slog.String("service", req.Name), slog.String("certificate", string(signedCert)),
			slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	authzPub, err := keygen.EncodePublic(&a.privateKey.PublicKey)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceRegistryFailed()

		a.logger.ErrorContext(ctx, "failed to encode public key",
			slog.String("service", req.Name), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SignUpResponse{
		Certificate: signedCert,
		Service: &pb.ID{
			PublicKey:   authzPub,
			Certificate: a.caCert,
		},
	}, nil
}

func (a *Authz) Register(ctx context.Context, req *pb.CertificateRequest) (*pb.CertificateResponse, error) {
	signUp, err := a.SignUp(ctx, &pb.SignUpRequest{
		Name:           req.Service,
		PublicKey:      req.PublicKey,
		SigningRequest: req.SigningRequest,
	})
	if err != nil {
		return nil, err
	}

	return &pb.CertificateResponse{
		Certificate: signUp.Certificate,
	}, nil
}
