package ca

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"log/slog"
	"slices"
	"time"

	"github.com/zalgonoise/cfg"
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
	ErrPrivateKey  = errs.Entity("private key")
	ErrCertificate = errs.Entity("certificate")
	ErrRepository  = errs.Entity("repository")
)

var (
	ErrNilRepository      = errs.WithDomain(errDomain, ErrNil, ErrRepository)
	ErrNilPrivateKey      = errs.WithDomain(errDomain, ErrNil, ErrPrivateKey)
	ErrInvalidPublicKey   = errs.WithDomain(errDomain, ErrInvalid, ErrPublicKey)
	ErrInvalidCertificate = errs.WithDomain(errDomain, ErrInvalid, ErrCertificate)
)

type Repository interface {
	Get(ctx context.Context, service string) (pubKey []byte, cert []byte, err error)
	Create(ctx context.Context, service string, pubKey []byte, cert []byte) (err error)
	Delete(ctx context.Context, service string) error
}

type CertificateAuthority struct {
	pb.UnimplementedCertificateAuthorityServer

	privateKey *ecdsa.PrivateKey
	ca         *x509.Certificate
	cert       *pem.Block
	durMonth   int

	repository Repository

	logger *slog.Logger
	tracer trace.Tracer
}

func NewCertificateAuthority(
	repo Repository,
	privateKey *ecdsa.PrivateKey,
	opts ...cfg.Option[Config],
) (*CertificateAuthority, error) {
	if repo == nil {
		return nil, ErrNilRepository
	}

	if privateKey == nil {
		return nil, ErrNilPrivateKey
	}

	config := cfg.New(opts...)

	if config.logHandler == nil {
		config.logHandler = log.NoOp{}
	}

	if config.tracer == nil {
		config.tracer = noop.NewTracerProvider().Tracer("x/authz/ca")
	}

	template := cfg.Set(newDefaultTemplate(), config.template...)

	ca, cert, err := NewCertificate(template)
	if err != nil {
		return nil, err
	}

	return &CertificateAuthority{
		privateKey: privateKey,
		ca:         ca,
		cert:       cert,
		durMonth:   template.DurMonth,
		repository: repo,
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

	block, _ := pem.Decode(req.PublicKey)
	if block == nil {
		return nil, status.Error(codes.InvalidArgument, "public key not in a PEM block")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	i, err := newInt(2, defaultExp, defaultSub)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	csr := toCSR(req.SigningRequest)

	if csr.Subject.CommonName == "" {
		csr.Subject.CommonName = req.Service
	}

	cert := &x509.Certificate{
		Version:         ca.ca.Version,
		SerialNumber:    i,
		Subject:         csr.Subject,
		Extensions:      csr.Extensions,
		ExtraExtensions: csr.ExtraExtensions,
		DNSNames:        csr.DNSNames,
		EmailAddresses:  csr.EmailAddresses,
		IPAddresses:     csr.IPAddresses,
		URIs:            csr.URIs,
		NotBefore:       time.Now(),
		NotAfter:        time.Now().AddDate(0, ca.durMonth, 0),
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageCodeSigning,
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageOCSPSigning,
			x509.ExtKeyUsageCodeSigning,
		},
		KeyUsage: x509.KeyUsageCertSign,
	}

	signedCertBytes, err := x509.CreateCertificate(rand.Reader, cert, ca.ca, pubKey, ca.privateKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	signedCert := pem.EncodeToMemory(&pem.Block{Type: typeCertificate, Bytes: signedCertBytes})

	if err := ca.repository.Create(ctx, req.Service, req.PublicKey, signedCert); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CertificateResponse{Certificate: signedCert}, nil
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
	key, err := x509.MarshalPKIXPublicKey(&ca.privateKey.PublicKey)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.PublicKeyResponse{PublicKey: key}, nil
}
