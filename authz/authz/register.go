package authz

import (
	"context"
	"crypto/ecdsa"
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

const (
	certificateLimit = 2
)

func (a *Authz) SignUp(ctx context.Context, req *pb.SignUpRequest) (*pb.SignUpResponse, error) {
	ctx, span := a.tracer.Start(ctx, "Authz.SignUp", trace.WithAttributes(
		attribute.String("service", req.Service),
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
		slog.String("service", req.Service), slog.Bool("with_csr", req.SigningRequest != nil))

	if err := req.ValidateAll(); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceRegistryFailed()

		a.logger.WarnContext(ctx, "invalid request",
			slog.Any("request", req), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	pubKey, err := keygen.DecodePublic(req.PublicKey)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncCertificatesDeleteFailed(req.Service)

		a.logger.WarnContext(ctx, "invalid public key",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, ErrInvalidPublicKey.Error())
	}

	if err = a.validatePublicKeys(ctx, req.Service, pubKey); err != nil {
		switch {
		// create the service if not exists
		case errors.Is(err, repository.ErrNotFound):
			if err := a.services.CreateService(ctx, req.Service, req.PublicKey); err != nil {
				span.SetStatus(otelcodes.Error, err.Error())
				span.RecordError(err)
				a.metrics.IncServiceRegistryFailed()

				a.logger.ErrorContext(ctx, "failed to write service to DB",
					slog.String("service", req.Service), slog.String("error", err.Error()))

				return nil, status.Error(codes.Internal, err.Error())
			}

		// service exists, invalid public keys
		case errors.Is(err, ErrInvalidPublicKey):
			span.SetStatus(otelcodes.Error, err.Error())
			span.RecordError(err)
			a.metrics.IncServiceRegistryFailed()

			a.logger.WarnContext(ctx, "mismatching public keys",
				slog.String("service", req.Service), slog.String("error", err.Error()))

			return nil, status.Error(codes.PermissionDenied, err.Error())

		// internal error
		default:
			span.SetStatus(otelcodes.Error, err.Error())
			span.RecordError(err)
			a.metrics.IncServiceRegistryFailed()

			a.logger.ErrorContext(ctx, "failed to validate public keys",
				slog.String("service", req.Service), slog.String("error", err.Error()))

			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	certRes, err := a.CreateCertificate(ctx, &pb.CertificateRequest{
		Service:        req.Service,
		PublicKey:      req.PublicKey,
		SigningRequest: req.SigningRequest,
	})
	if err != nil {
		return nil, err
	}

	return &pb.SignUpResponse{
		Certificate:        certRes.Certificate,
		ServiceCertificate: a.certRaw,
	}, nil
}

func (a *Authz) CreateCertificate(ctx context.Context, req *pb.CertificateRequest) (*pb.CertificateResponse, error) {
	ctx, span := a.tracer.Start(ctx, "Authz.CreateCertificate", trace.WithAttributes(
		attribute.String("service", req.Service),
		attribute.String("pub_key", string(req.PublicKey)),
		attribute.Bool("with_csr", req.SigningRequest != nil),
	))
	defer span.End()

	start := time.Now()
	defer func() {
		a.metrics.ObserveCertificatesCreateLatency(ctx, req.Service, time.Since(start))
	}()

	a.metrics.IncCertificatesCreated(req.Service)
	a.logger.DebugContext(ctx, "new certificate creation request",
		slog.String("service", req.Service), slog.Bool("with_csr", req.SigningRequest != nil))

	if err := req.ValidateAll(); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncCertificatesCreateFailed(req.Service)

		a.logger.WarnContext(ctx, "invalid request",
			slog.Any("request", req), slog.String("error", err.Error()))

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	certificates, err := a.services.ListCertificates(ctx, req.Service)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceRegistryFailed()
		a.metrics.IncCertificatesCreateFailed(req.Service)

		a.logger.ErrorContext(ctx, "failed to get service certificates from DB",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	// max number of certificates, return the stored certificate with the biggest validity
	if len(certificates) >= certificateLimit {
		return certificates[0], nil
	}

	certificate, expiry, err := a.newCertificate(ctx, req)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceRegistryFailed()
		a.metrics.IncCertificatesCreateFailed(req.Service)

		if errors.Is(err, ErrInvalidPublicKey) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := a.services.CreateCertificate(ctx, req.Service, certificate, expiry); err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceRegistryFailed()
		a.metrics.IncCertificatesCreateFailed(req.Service)

		a.logger.ErrorContext(ctx, "failed to write certificate to DB",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CertificateResponse{Certificate: certificate}, nil
}

func (a *Authz) Register(ctx context.Context, req *pb.CertificateRequest) (*pb.CertificateResponse, error) {
	signUp, err := a.SignUp(ctx, &pb.SignUpRequest{
		Service:        req.Service,
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

func (a *Authz) newCertificate(ctx context.Context, req *pb.CertificateRequest) ([]byte, time.Time, error) {
	pubKey, err := keygen.DecodePublic(req.PublicKey)
	if err != nil {
		a.logger.WarnContext(ctx, "invalid public key",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, time.Time{}, ErrInvalidPublicKey
	}

	csr := certs.ToCSR(req.Service, pubKey, req.SigningRequest)
	cert, err := certs.NewCertFromCSR(a.cert.Version, a.durMonth, csr)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to generate new certificate",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, time.Time{}, err
	}

	signedCert, err := certs.Encode(cert, a.cert, pubKey, a.privateKey)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to encode the new certificate",
			slog.String("service", req.Service), slog.String("error", err.Error()))

		return nil, time.Time{}, err
	}

	return signedCert, cert.NotAfter, nil
}

func (a *Authz) validatePublicKeys(ctx context.Context, service string, key *ecdsa.PublicKey) error {
	storedPub, err := a.services.GetService(ctx, service)
	if err != nil {
		return err
	}

	pub, err := keygen.DecodePublic(storedPub)
	if err != nil {
		return err
	}

	if !pub.Equal(key) {
		return ErrInvalidPublicKey
	}

	return nil
}
