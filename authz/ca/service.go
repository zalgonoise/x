package ca

import (
	"context"
	"crypto/ecdsa"
	"encoding/pem"
	"errors"
	"log/slog"
	"slices"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/authz/keygen"
	"github.com/zalgonoise/x/authz/log"
	pb "github.com/zalgonoise/x/authz/pb/authz/v1"
	"github.com/zalgonoise/x/authz/repository"
	"github.com/zalgonoise/x/errs"
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
	ErrCertificate = errs.Entity("certificate")
	ErrRepository  = errs.Entity("repository")
	ErrVerifier    = errs.Entity("verifier")
	ErrSigner      = errs.Entity("signer")
)

var (
	ErrNilRepository      = errs.WithDomain(errDomain, ErrNil, ErrRepository)
	ErrNilVerifier        = errs.WithDomain(errDomain, ErrNil, ErrVerifier)
	ErrNilSigner          = errs.WithDomain(errDomain, ErrNil, ErrSigner)
	ErrInvalidPublicKey   = errs.WithDomain(errDomain, ErrInvalid, ErrPublicKey)
	ErrInvalidCertificate = errs.WithDomain(errDomain, ErrInvalid, ErrCertificate)
)

type Repository interface {
	Get(ctx context.Context, service string) (pubKey []byte, cert []byte, err error)
	Create(ctx context.Context, service string, pubKey []byte, cert []byte) (err error)
	Delete(ctx context.Context, service string) error
}

type Verifier interface {
	Verify(data []byte, signature []byte) error
	Hash(data []byte) (hash []byte, err error)
	Key() ecdsa.PublicKey
}

type Signer interface {
	Sign(data []byte) (signature []byte, err error)
	Hash(data []byte) (hash []byte, err error)
	Key() ecdsa.PublicKey
}

type CertificateAuthority struct {
	pb.UnimplementedCertificateAuthorityServer

	pubKey ecdsa.PublicKey
	cert   *pem.Block

	repository Repository
	verifier   Verifier
	signer     Signer

	logger *slog.Logger
	tracer trace.Tracer
}

func NewCertificateAuthority(
	repo Repository,
	verifier Verifier,
	signer Signer,
	opts ...cfg.Option[Config],
) (*CertificateAuthority, error) {
	if repo == nil {
		return nil, ErrNilRepository
	}

	if verifier == nil {
		return nil, ErrNilVerifier
	}

	if signer == nil {
		return nil, ErrNilSigner
	}

	config := cfg.New(opts...)

	if config.logHandler == nil {
		config.logHandler = log.NoOp{}
	}

	if config.tracer == nil {
		config.tracer = noop.NewTracerProvider().Tracer("x/authz/ca")
	}

	cert, err := NewCertificate(config.template...)
	if err != nil {
		return nil, err
	}

	return &CertificateAuthority{
		pubKey:     signer.Key(),
		cert:       cert,
		repository: repo,
		verifier:   verifier,
		signer:     signer,
		logger:     slog.New(config.logHandler),
		tracer:     config.tracer,
	}, nil
}

func (ca *CertificateAuthority) Register(
	ctx context.Context, req *pb.CertificateRequest) (*pb.CertificateResponse, error) {
	_, _, err := ca.repository.Get(ctx, req.Service)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = keygen.DecodePublic(req.PublicKey)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// TODO: swap this basic pubkey signing with an actual x509 certificate from the configured template
	signature, err := ca.signer.Sign(req.PublicKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := ca.repository.Create(ctx, req.Service, req.PublicKey, signature); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CertificateResponse{Certificate: signature}, nil
}

func (ca *CertificateAuthority) GetCertificate(ctx context.Context, req *pb.CertificateRequest) (*pb.CertificateResponse, error) {
	pubKey, cert, err := ca.repository.Get(ctx, req.Service)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	if !slices.Equal(pubKey, req.PublicKey) {
		return nil, status.Error(codes.PermissionDenied, ErrInvalidPublicKey.Error())
	}

	return &pb.CertificateResponse{Certificate: cert}, nil
}

func (ca *CertificateAuthority) DeleteService(ctx context.Context, req *pb.DeletionRequest) (*pb.DeletionResponse, error) {
	pubKey, cert, err := ca.repository.Get(ctx, req.Service)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	if !slices.Equal(pubKey, req.PublicKey) {
		return nil, status.Error(codes.PermissionDenied, ErrInvalidPublicKey.Error())
	}

	if !slices.Equal(cert, req.Certificate) {
		return nil, status.Error(codes.PermissionDenied, ErrInvalidCertificate.Error())
	}

	if err = ca.repository.Delete(ctx, req.Service); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeletionResponse{}, nil
}

func (ca *CertificateAuthority) PublicKey(context.Context, *pb.PublicKeyRequest) (*pb.PublicKeyResponse, error) {
	key, err := keygen.EncodePublic(&ca.pubKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.PublicKeyResponse{PublicKey: key}, nil
}
