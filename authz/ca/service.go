package ca

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"log/slog"
	"slices"

	"github.com/zalgonoise/x/authz/keygen"
	pb "github.com/zalgonoise/x/authz/pb/authz/v1"
	"github.com/zalgonoise/x/authz/repository"
	"github.com/zalgonoise/x/errs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	errDomain = errs.Domain("x/authz/ca")

	ErrNil     = errs.Kind("nil")
	ErrInvalid = errs.Kind("invalid")

	ErrPrivateKey  = errs.Entity("private key")
	ErrPublicKey   = errs.Entity("public key")
	ErrCertificate = errs.Entity("certificate")
	ErrRepository  = errs.Entity("repository")
	ErrSigner      = errs.Entity("signer")
)

var (
	ErrNilPrivateKey      = errs.WithDomain(errDomain, ErrNil, ErrPrivateKey)
	ErrNilRepository      = errs.WithDomain(errDomain, ErrNil, ErrRepository)
	ErrNilSigner          = errs.WithDomain(errDomain, ErrNil, ErrSigner)
	ErrInvalidPublicKey   = errs.WithDomain(errDomain, ErrInvalid, ErrPublicKey)
	ErrInvalidCertificate = errs.WithDomain(errDomain, ErrInvalid, ErrCertificate)
)

type Repository interface {
	Get(ctx context.Context, service string) (pubKey []byte, cert []byte, err error)
	Create(ctx context.Context, service string, pubKey []byte, cert []byte) (err error)
	Delete(ctx context.Context, service string) error
}

type Signer interface {
	Sign(data []byte, privateKey *ecdsa.PrivateKey) ([]byte, error)
	Verify(data []byte, pubKey *ecdsa.PublicKey, signature []byte) error
}

type CertificateAuthority struct {
	pb.UnimplementedCertificateAuthorityServer

	privateKey *ecdsa.PrivateKey

	repository Repository
	signer     Signer

	logger *slog.Logger
}

func NewCertificateAuthority(
	privateKey *ecdsa.PrivateKey,
	repo Repository,
	signer Signer,
	logger *slog.Logger,
) (*CertificateAuthority, error) {
	if privateKey == nil {
		return nil, ErrNilPrivateKey
	}

	if repo == nil {
		return nil, ErrNilRepository
	}

	if signer == nil {
		return nil, ErrNilSigner
	}

	return &CertificateAuthority{
		privateKey: privateKey,
		repository: repo,
		signer:     signer,
		logger:     logger,
	}, nil
}

func (ca *CertificateAuthority) Register(
	ctx context.Context, req *pb.CertificateRequest) (*pb.CertificateResponse, error) {
	_, _, err := ca.repository.Get(ctx, req.Service)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	signature, err := ca.signer.Sign(req.PublicKey, ca.privateKey)
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
	key, err := keygen.EncodePublic(&ca.privateKey.PublicKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.PublicKeyResponse{PublicKey: key}, nil

}
