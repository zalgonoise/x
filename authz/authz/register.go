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
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
)

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

	exit := withExit[pb.SignUpRequest, pb.SignUpResponse](
		ctx, a.logger, req, a.metrics.IncServiceRegistryFailed, span,
	)

	if err := req.ValidateAll(); err != nil {
		return exit(codes.InvalidArgument, "invalid request", err)
	}

	storedPubKey, storedCert, err := a.services.GetService(ctx, req.Name)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return exit(codes.Internal, "failed to get service from DB", ErrInvalidService)
	}

	// service is already registered, just return the stored certificate
	if err == nil && len(storedPubKey) > 0 && len(storedCert) > 0 {
		authzPub, err := keygen.EncodePublic(&a.privateKey.PublicKey)
		if err != nil {
			return exit(codes.Internal, "failed to encode public key", err)
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
		return exit(codes.InvalidArgument, "invalid request", err)
	}

	cert, err := certs.NewCertFromCSR(a.cert.Version, a.durMonth, certs.ToCSR(req.Name, pubKey, req.SigningRequest))
	if err != nil {
		return exit(codes.Internal, "failed to generate new serial number", err)
	}

	signedCert, err := certs.Encode(cert, a.cert, pubKey, a.privateKey)
	if err != nil {
		return exit(codes.Internal, "failed to generate new certificate", err)
	}

	if err := a.services.CreateService(ctx, req.Name, req.PublicKey, signedCert, cert.NotAfter); err != nil {
		return exit(codes.Internal, "failed to write certificate to DB", err,
			slog.String("certificate", string(signedCert)),
		)
	}

	authzPub, err := keygen.EncodePublic(&a.privateKey.PublicKey)
	if err != nil {
		return exit(codes.Internal, "failed to encode public key", err)
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
