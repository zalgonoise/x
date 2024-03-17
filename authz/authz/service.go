package authz

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"log/slog"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/authz/ca"
	"github.com/zalgonoise/x/authz/certs"
	"github.com/zalgonoise/x/authz/keygen"
	pb "github.com/zalgonoise/x/authz/pb/authz/v1"
	"github.com/zalgonoise/x/authz/randomizer"
	"github.com/zalgonoise/x/authz/repository"
	"github.com/zalgonoise/x/errs"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultConnTimeout     = time.Minute
	defaultChallengeSize   = 32
	defaultDurMonth        = 12
	defaultChallengeExpiry = 10 * time.Minute
	defaultTokenExpiry     = time.Hour
)

const (
	errDomain = errs.Domain("x/authz/authz")

	ErrNil     = errs.Kind("nil")
	ErrInvalid = errs.Kind("invalid")
	ErrExpired = errs.Kind("expired")
	ErrEmpty   = errs.Kind("empty")

	ErrCAAddress          = errs.Entity("CA address")
	ErrPublicKey          = errs.Entity("public key")
	ErrServicePublicKey   = errs.Entity("service public key")
	ErrIDPublicKey        = errs.Entity("ID public key")
	ErrPrivateKey         = errs.Entity("private key")
	ErrCertificate        = errs.Entity("certificate")
	ErrServiceCertificate = errs.Entity("service certificate")
	ErrIDCertificate      = errs.Entity("ID certificate")
	ErrServicesRepo       = errs.Entity("services repository")
	ErrTokensRepo         = errs.Entity("tokens repository")
	ErrSignature          = errs.Entity("signature")
	ErrChallenge          = errs.Entity("challenge")
	ErrToken              = errs.Entity("token")
	ErrService            = errs.Entity("service")
)

var (
	ErrEmptyCAAddress            = errs.WithDomain(errDomain, ErrEmpty, ErrCAAddress)
	ErrNilServicesRepository     = errs.WithDomain(errDomain, ErrNil, ErrServicesRepo)
	ErrNilTokensRepository       = errs.WithDomain(errDomain, ErrNil, ErrTokensRepo)
	ErrNilPrivateKey             = errs.WithDomain(errDomain, ErrNil, ErrPrivateKey)
	ErrInvalidPublicKey          = errs.WithDomain(errDomain, ErrInvalid, ErrPublicKey)
	ErrInvalidServicePublicKey   = errs.WithDomain(errDomain, ErrInvalid, ErrServicePublicKey)
	ErrInvalidServiceCertificate = errs.WithDomain(errDomain, ErrInvalid, ErrServiceCertificate)
	ErrInvalidIDPublicKey        = errs.WithDomain(errDomain, ErrInvalid, ErrIDPublicKey)
	ErrInvalidIDCertificate      = errs.WithDomain(errDomain, ErrInvalid, ErrIDCertificate)
	ErrInvalidSignature          = errs.WithDomain(errDomain, ErrInvalid, ErrSignature)
	ErrInvalidChallenge          = errs.WithDomain(errDomain, ErrInvalid, ErrChallenge)
	ErrInvalidService            = errs.WithDomain(errDomain, ErrInvalid, ErrService)
)

type ServiceRepository interface {
	CreateService(ctx context.Context, service string, pubKey []byte, cert []byte, expiry time.Time) (err error)
	GetService(ctx context.Context, service string) (pubKey []byte, cert []byte, err error)
	DeleteService(ctx context.Context, service string) error
}

type TokensRepository interface {
	CreateChallenge(ctx context.Context, service string, challenge []byte, expiry time.Time) error
	ListChallenges(ctx context.Context, service string) (challenges []repository.Challenge, err error)
	DeleteChallenge(ctx context.Context, service string, challenge []byte) error

	CreateToken(ctx context.Context, service string, token []byte, expiry time.Time) error
	ListTokens(ctx context.Context, service string) (tokens []repository.Token, err error)
	DeleteToken(ctx context.Context, service string, token []byte) error
}

type Randomizer interface {
	Random() ([]byte, error)
}

type Metrics interface {
	// Authz metrics
	IncServiceLoginRequests(service string)
	IncServiceLoginFailed(service string)
	ObserveServiceLoginLatency(ctx context.Context, service string, duration time.Duration)
	IncServiceTokenRequests(service string)
	IncServiceTokenFailed(service string)
	ObserveServiceTokenLatency(ctx context.Context, service string, duration time.Duration)
	IncServiceTokenVerifications(service string)
	IncServiceTokenVerificationFailed(service string)
	ObserveServiceTokenVerificationLatency(ctx context.Context, service string, duration time.Duration)

	// CA metrics
	IncServiceRegistries()
	IncServiceRegistryFailed()
	ObserveServiceRegistryLatency(ctx context.Context, duration time.Duration)
	IncServiceCertsFetched(service string)
	IncServiceCertsFetchFailed(service string)
	ObserveServiceCertsFetchLatency(ctx context.Context, service string, duration time.Duration)
	IncVerificationRequests(service string)
	IncVerificationFailed(service string)
	ObserveVerificationLatency(ctx context.Context, service string, duration time.Duration)
	IncServiceDeletions()
	IncServiceDeletionFailed()
	ObserveServiceDeletionLatency(ctx context.Context, duration time.Duration)
	IncPubKeyRequests()
	IncPubKeyRequestFailed()
	ObservePubKeyRequestLatency(ctx context.Context, duration time.Duration)
	RegisterCollector(collector prometheus.Collector)
}

type Authz struct {
	pb.UnimplementedAuthzServer
	pb.UnimplementedCertificateAuthorityServer

	caClient pb.CertificateAuthorityClient

	name       string
	privateKey *ecdsa.PrivateKey
	cert       *x509.Certificate
	caCert     []byte
	durMonth   int

	challengeExpiry time.Duration
	tokenExpiry     time.Duration

	ca       *ca.CertificateAuthority
	services ServiceRepository
	tokens   TokensRepository
	random   Randomizer

	metrics Metrics
	logger  *slog.Logger
	tracer  trace.Tracer
}

func NewAuthz(
	name, caAddress string,
	privateKey *ecdsa.PrivateKey,
	services ServiceRepository,
	tokens TokensRepository,
	random Randomizer,
	opts ...cfg.Option[Config],
) (*Authz, error) {
	if caAddress == "" {
		return nil, ErrEmptyCAAddress
	}

	if privateKey == nil {
		return nil, ErrNilPrivateKey
	}

	if services == nil {
		return nil, ErrNilServicesRepository
	}

	if tokens == nil {
		return nil, ErrNilTokensRepository
	}

	if random == nil {
		random = randomizer.New(defaultChallengeSize)
	}

	config := cfg.Set(defaultConfig(), opts...)

	conn, err := dial(caAddress, config.m)
	if err != nil {
		return nil, err
	}

	logger := slog.New(config.logger)
	logger.DebugContext(context.Background(), "connected to CA")

	caClient := pb.NewCertificateAuthorityClient(conn)

	pub, err := keygen.EncodePublic(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}

	ctx, done := context.WithTimeout(context.Background(), defaultConnTimeout)
	defer done()

	res, err := caClient.Register(ctx, &pb.CertificateRequest{
		Service:        name,
		PublicKey:      pub,
		SigningRequest: config.csr,
	})
	if err != nil {
		return nil, err
	}

	logger.InfoContext(ctx, "retrieved certificate from CA")

	cert, err := certs.Decode(res.Certificate)
	if err != nil {
		return nil, err
	}

	return &Authz{
		caClient:        caClient,
		name:            name,
		privateKey:      privateKey,
		cert:            cert,
		caCert:          res.Certificate,
		durMonth:        config.durMonth,
		challengeExpiry: config.challengeExpiry,
		tokenExpiry:     config.tokenExpiry,
		services:        services,
		tokens:          tokens,
		random:          random,
		metrics:         config.m,
		logger:          logger,
		tracer:          config.tracer,
	}, nil
}

func dial(uri string, m Metrics) (*grpc.ClientConn, error) {
	clientMetrics := grpc_prometheus.NewClientMetrics(grpc_prometheus.WithClientHandlingTimeHistogram())

	conn, err := grpc.Dial(uri,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(
			grpc_middleware.ChainUnaryClient(
				clientMetrics.UnaryClientInterceptor(),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	m.RegisterCollector(clientMetrics)

	return conn, nil
}
