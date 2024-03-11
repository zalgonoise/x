package authz

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha512"
	"crypto/x509"
	"errors"
	"log/slog"
	"slices"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/authz/ca"
	"github.com/zalgonoise/x/authz/certs"
	"github.com/zalgonoise/x/authz/keygen"
	"github.com/zalgonoise/x/authz/randomizer"
	"github.com/zalgonoise/x/authz/repository"
	"github.com/zalgonoise/x/errs"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	pb "github.com/zalgonoise/x/authz/pb/authz/v1"
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
	ErrInvalidCertificate        = errs.WithDomain(errDomain, ErrInvalid, ErrCertificate)
	ErrInvalidServicePublicKey   = errs.WithDomain(errDomain, ErrInvalid, ErrServicePublicKey)
	ErrInvalidServiceCertificate = errs.WithDomain(errDomain, ErrInvalid, ErrServiceCertificate)
	ErrInvalidIDPublicKey        = errs.WithDomain(errDomain, ErrInvalid, ErrIDPublicKey)
	ErrInvalidIDCertificate      = errs.WithDomain(errDomain, ErrInvalid, ErrIDCertificate)
	ErrInvalidSignature          = errs.WithDomain(errDomain, ErrInvalid, ErrSignature)
	ErrExpiredChallenge          = errs.WithDomain(errDomain, ErrExpired, ErrChallenge)
	ErrExpiredToken              = errs.WithDomain(errDomain, ErrExpired, ErrToken)
	ErrInvalidService            = errs.WithDomain(errDomain, ErrInvalid, ErrService)
)

type ServiceRepository interface {
	CreateService(ctx context.Context, service string, pubKey []byte, cert []byte, expiry time.Time) (err error)
	GetService(ctx context.Context, service string) (pubKey []byte, cert []byte, err error)
	DeleteService(ctx context.Context, service string) error
}

type TokensRepository interface {
	CreateChallenge(ctx context.Context, service string, challenge []byte, expiry time.Time) error
	GetChallenge(ctx context.Context, service string) (challenge []byte, expiry time.Time, err error)
	DeleteChallenge(ctx context.Context, service string) error

	CreateToken(ctx context.Context, service string, token []byte, expiry time.Time) error
	GetToken(ctx context.Context, service string) (token []byte, expiry time.Time, err error)
	DeleteToken(ctx context.Context, service string) error
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

func (a *Authz) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	ctx, span := a.tracer.Start(ctx, "Authz.Login", trace.WithAttributes(
		attribute.String("service", req.Name),
		attribute.String("id.pub_key", string(req.Id.PublicKey)),
	))
	defer span.End()

	start := time.Now()
	defer func() {
		a.metrics.ObserveServiceLoginLatency(ctx, req.Name, time.Since(start))
	}()

	a.metrics.IncServiceLoginRequests(req.Name)
	a.logger.DebugContext(ctx, "new login request", slog.String("service", req.Name))

	exit := withExit[pb.LoginRequest, pb.LoginResponse](
		ctx, a.logger, req, func() { a.metrics.IncServiceLoginFailed(req.Name) }, span,
	)

	if err := req.ValidateAll(); err != nil {
		return exit(codes.InvalidArgument, "invalid request", err)
	}

	storedPEM, _, err := a.services.GetService(ctx, req.Name)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return exit(codes.InvalidArgument, "couldn't find service in the database", ErrInvalidService)
		}

		return exit(codes.Internal, "failed to fetch stored public key", err)
	}

	servicePubKey, err := keygen.DecodePublic(req.Service.PublicKey)
	if err != nil {
		return exit(codes.InvalidArgument, "invalid service public key PEM bytes", err)
	}

	if !a.privateKey.PublicKey.Equal(servicePubKey) {
		return exit(codes.InvalidArgument, "mismatching service public keys", ErrInvalidServicePublicKey)
	}

	serviceCert, err := certs.Decode(req.Service.Certificate)
	if err != nil {
		return exit(codes.InvalidArgument, "invalid service cert bytes", err)
	}

	if !a.cert.Equal(serviceCert) {
		return exit(codes.InvalidArgument, "mismatching service cert", ErrInvalidServiceCertificate)
	}

	pubKey, err := keygen.DecodePublic(req.Id.PublicKey)
	if err != nil {
		return exit(codes.InvalidArgument, "invalid public key PEM bytes", err)
	}

	cert, err := certs.Decode(req.Id.Certificate)
	if err != nil {
		return exit(codes.InvalidArgument, "invalid cert bytes", err)
	}

	pub, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return exit(codes.InvalidArgument, "failed to retrieve public key from certificate", ErrInvalidIDCertificate)
	}

	storedPub, err := keygen.DecodePublic(storedPEM)
	if err != nil {
		return exit(codes.Internal, "failed to decode stored public key", err)
	}

	if !pub.Equal(storedPub) {
		return exit(codes.PermissionDenied, "mismatching public keys", ErrInvalidIDPublicKey)
	}

	if !pubKey.Equal(storedPub) {
		return exit(codes.PermissionDenied, "mismatching public keys", ErrInvalidIDPublicKey)
	}

	if time.Now().After(cert.NotAfter) {
		a.logger.DebugContext(ctx, "expired certificate",
			slog.Time("expiry", cert.NotAfter), slog.String("service", req.Name))

		return exit(codes.InvalidArgument, "expired certificate", ErrInvalidIDCertificate)
	}

	// check if there is a valid challenge to provide
	challenge, expiry, err := a.tokens.GetChallenge(ctx, req.Name)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return exit(codes.Internal, "failed to search for challenges in the database", err)
	}

	if err == nil && len(challenge) > 0 {
		if expiry.After(start) {
			return &pb.LoginResponse{Challenge: challenge, ExpiresOn: expiry.UnixMilli()}, nil
		}

		if err = a.tokens.DeleteChallenge(ctx, req.Name); err != nil {
			return exit(codes.Internal, "failed to remove expired challenge", err)
		}
	}

	// create a new challenge
	challenge, err = a.random.Random()
	if err != nil {
		return exit(codes.Internal, "failed to generate challenge", err)
	}

	expiry = time.Now().Add(a.challengeExpiry)

	if err = a.tokens.CreateChallenge(ctx, req.Name, challenge, expiry); err != nil {
		return exit(codes.Internal, "failed to store challenge", err)
	}

	return &pb.LoginResponse{Challenge: challenge, ExpiresOn: expiry.UnixMilli()}, nil
}

func (a *Authz) Token(ctx context.Context, req *pb.TokenRequest) (*pb.TokenResponse, error) {
	ctx, span := a.tracer.Start(ctx, "Authz.Token", trace.WithAttributes(
		attribute.String("service", req.Name),
	))
	defer span.End()

	start := time.Now()
	defer func() {
		a.metrics.ObserveServiceTokenLatency(ctx, req.Name, time.Since(start))
	}()

	a.metrics.IncServiceTokenRequests(req.Name)
	a.logger.DebugContext(ctx, "new certificate request",
		slog.String("service", req.Name))

	exit := withExit[pb.TokenRequest, pb.TokenResponse](
		ctx, a.logger, req, func() { a.metrics.IncServiceTokenFailed(req.Name) }, span,
	)

	if err := req.ValidateAll(); err != nil {
		return exit(codes.InvalidArgument, "invalid request", err)
	}

	storedPub, _, err := a.services.GetService(ctx, req.Name)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return exit(codes.InvalidArgument, "service does not exist", err)
		}

		return exit(codes.Internal, "failed to get service details", err)
	}

	challenge, expiry, err := a.tokens.GetChallenge(ctx, req.Name)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return exit(codes.InvalidArgument, "couldn't find a challenge for this token request", err)
		}

		return exit(codes.Internal, "failed to get challenge", err)
	}

	if time.Now().After(expiry) {
		return exit(codes.InvalidArgument, "challenge is expired", ErrExpiredChallenge)
	}

	pub, err := keygen.DecodePublic(storedPub)
	if err != nil {
		return exit(codes.Internal, "failed to decode public key", err)
	}

	h := sha512.Sum512(challenge)

	if !ecdsa.VerifyASN1(pub, h[:], req.SignedChallenge) {
		return exit(codes.InvalidArgument, "invalid signature", ErrInvalidSignature)
	}

	exp := time.Now().Add(a.tokenExpiry)
	token, err := keygen.NewToken(a.privateKey, a.name, exp, keygen.WithClaim(keygen.Claim{
		Service: req.Name,
		Authz:   a.name,
	}))
	if err != nil {
		return exit(codes.Internal, "failed to generate JWT", err)
	}

	if err = a.tokens.CreateToken(ctx, req.Name, token, exp); err != nil {
		return exit(codes.Internal, "failed to store token", err)
	}

	return &pb.TokenResponse{
		Token:     string(token),
		ExpiresOn: exp.UnixMilli(),
	}, nil
}

func (a *Authz) VerifyToken(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	ctx, span := a.tracer.Start(ctx, "Authz.VerifyToken")
	defer span.End()

	token, err := keygen.ParseToken([]byte(req.Token), &a.privateKey.PublicKey)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
		a.metrics.IncServiceTokenVerificationFailed("")
		a.logger.WarnContext(ctx, "failed to decode token")

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	start := time.Now()
	defer func() {
		a.metrics.ObserveServiceTokenVerificationLatency(ctx, token.Claim.Service, time.Since(start))
	}()

	a.metrics.IncServiceTokenVerifications(token.Claim.Service)
	a.logger.DebugContext(ctx, "new token verification request")

	exit := withExit[pb.AuthRequest, pb.AuthResponse](
		ctx, a.logger, req, func() {
			a.metrics.IncServiceTokenVerificationFailed(token.Claim.Service)
		}, span,
	)

	if err := req.ValidateAll(); err != nil {
		return exit(codes.InvalidArgument, "invalid request", err)
	}

	storedToken, exp, err := a.tokens.GetToken(ctx, token.Claim.Service)
	if err != nil {
		return exit(codes.InvalidArgument, "invalid token", err)
	}

	if string(storedToken) != req.Token {
		return exit(codes.InvalidArgument, "invalid token", err)
	}

	if !exp.Equal(token.Expiry) {
		return exit(codes.InvalidArgument, "invalid token", err)
	}

	if start.After(exp) {
		if err = a.tokens.DeleteToken(ctx, token.Claim.Service); err != nil {
			a.logger.WarnContext(ctx, "failed to remove expired token",
				slog.String("error", err.Error()),
				slog.String("service", token.Claim.Service),
			)
		}

		return exit(codes.PermissionDenied, "token is expired", ErrExpiredToken)
	}

	return &pb.AuthResponse{}, nil
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

func withExit[Req any, Res any](
	ctx context.Context, logger *slog.Logger,
	req *Req, metric func(), span trace.Span,
) func(codes.Code, string, error, ...any) (*Res, error) {
	return func(code codes.Code, message string, err error, args ...any) (*Res, error) {
		logArgs := make([]any, 0, len(args)+2)

		if req != nil {
			logArgs = append(logArgs, slog.Any("request", req))
		}

		if err != nil {
			logArgs = append(logArgs, slog.String("error", err.Error()))
			logArgs = append(logArgs, args...)

			span.SetStatus(otelcodes.Error, err.Error())
			span.RecordError(err)
			metric()
			logger.WarnContext(ctx, message, logArgs...)

			return nil, status.Error(code, err.Error())
		}

		logArgs = append(logArgs, args...)
		span.SetStatus(otelcodes.Error, message)
		span.RecordError(errors.New(message))
		metric()
		logger.WarnContext(ctx, message, logArgs...)

		return nil, status.Error(code, message)
	}
}
