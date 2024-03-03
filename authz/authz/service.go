package authz

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
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
	"github.com/zalgonoise/x/authz/keygen"
	"github.com/zalgonoise/x/authz/repository"
	"github.com/zalgonoise/x/errs"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/zalgonoise/x/authz/log"
	"github.com/zalgonoise/x/authz/metrics"
	pb "github.com/zalgonoise/x/authz/pb/authz/v1"
)

const defaultConnTimeout = time.Minute

const (
	errDomain = errs.Domain("x/authz/authz")

	ErrNil     = errs.Kind("nil")
	ErrInvalid = errs.Kind("invalid")

	ErrPublicKey   = errs.Entity("public key")
	ErrPrivateKey  = errs.Entity("private key")
	ErrCertificate = errs.Entity("certificate")
	ErrRepository  = errs.Entity("repository")
	ErrSignature   = errs.Entity("signature")
)

var (
	ErrNilRepository      = errs.WithDomain(errDomain, ErrNil, ErrRepository)
	ErrNilPrivateKey      = errs.WithDomain(errDomain, ErrNil, ErrPrivateKey)
	ErrInvalidPublicKey   = errs.WithDomain(errDomain, ErrInvalid, ErrPublicKey)
	ErrInvalidCertificate = errs.WithDomain(errDomain, ErrInvalid, ErrCertificate)
	ErrInvalidSignature   = errs.WithDomain(errDomain, ErrInvalid, ErrSignature)
)

type ServiceRepository interface {
	GetService(ctx context.Context, service string) (pubKey []byte, cert []byte, err error)
	CreateService(ctx context.Context, service string, pubKey []byte, cert []byte, expiry time.Time) (err error)
	DeleteService(ctx context.Context, service string) error
}

type ChallengeRepository interface {
	CreateChallenge(ctx context.Context, service string, challenge []byte) error
	GetChallenge(ctx context.Context, service string) (challenge []byte, err error)
	DeleteChallenge(ctx context.Context, service string) error
}

type Randomizer interface {
	Random() []byte
}

type Metrics interface {
	// Authz metrics
	IncServiceLoginRequests(service string)
	IncServiceLoginFailed(service string)
	ObserveServiceLoginLatency(ctx context.Context, service string, duration time.Duration)
	IncServiceTokenRequests(service string)
	IncServiceTokenFailed(service string)
	ObserveServiceTokenLatency(ctx context.Context, service string, duration time.Duration)

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

	privateKey *ecdsa.PrivateKey
	cert       *x509.Certificate
	caCert     []byte
	durMonth   int

	ca         *ca.CertificateAuthority
	services   ServiceRepository
	challenges ChallengeRepository
	random     Randomizer

	metrics Metrics
	logger  *slog.Logger
	tracer  trace.Tracer
}

func NewAuthz(
	name, caAddress string,
	privateKey *ecdsa.PrivateKey,
	services ServiceRepository,
	challenges ChallengeRepository,
	randomizer Randomizer,
	opts ...cfg.Option[Config],
) (*Authz, error) {
	config := cfg.New(opts...)

	if config.m == nil {
		config.m = metrics.NoOp()
	}

	if config.logger == nil {
		config.logger = log.New("info")
	}

	if config.tracer == nil {
		config.tracer = noop.NewTracerProvider().Tracer("authz")
	}

	conn, err := dial(caAddress, config.m)
	if err != nil {
		return nil, err
	}

	config.logger.DebugContext(context.Background(), "connected to CA")

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

	config.logger.InfoContext(ctx, "retrieved certificate from CA")

	cert, err := keygen.DecodeCertificate(res.Certificate)
	if err != nil {
		return nil, err
	}

	return &Authz{
		caClient:   caClient,
		privateKey: privateKey,
		cert:       cert,
		caCert:     res.Certificate,
		services:   services,
		challenges: challenges,
		random:     randomizer,
		metrics:    config.m,
		logger:     config.logger,
		tracer:     config.tracer,
	}, nil
}

func (a *Authz) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	ctx, span := a.tracer.Start(ctx, "Authz.Register", trace.WithAttributes(
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
	a.logger.DebugContext(ctx, "new registry request",
		slog.String("service", req.Name), slog.Bool("with_csr", req.SigningRequest != nil))

	exit := withExit[pb.RegisterRequest, pb.RegisterResponse](
		ctx, a.logger, req, a.metrics.IncServiceRegistryFailed, span,
	)

	if err := req.ValidateAll(); err != nil {
		return exit(codes.InvalidArgument, "invalid request", err)
	}

	storedPubKey, storedCert, err := a.services.GetService(ctx, req.Name)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return exit(codes.Internal, "failed to get service from DB", err)
	}

	// service is already registered, just return the stored certificate
	if err == nil && len(storedPubKey) > 0 && len(storedCert) > 0 {
		authzPub, err := keygen.EncodePublic(&a.privateKey.PublicKey)
		if err != nil {
			return exit(codes.Internal, "failed to encode public key", err)
		}

		return &pb.RegisterResponse{
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

	cert, err := keygen.NewCertFromCSR(a.cert.Version, a.durMonth, keygen.ToCSR(req.Name, pubKey, req.SigningRequest))
	if err != nil {
		return exit(codes.Internal, "failed to generate new serial number", err)
	}

	signedCert, err := keygen.EncodeCertificate(cert, a.cert, pubKey, a.privateKey)
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

	return &pb.RegisterResponse{
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

	servicePubKey, err := keygen.DecodePublic(req.Service.PublicKey)
	if err != nil {
		return exit(codes.InvalidArgument, "invalid service public key PEM bytes", err)
	}

	if !a.privateKey.PublicKey.Equal(servicePubKey) {
		return exit(codes.InvalidArgument, "mismatching service public keys", ErrInvalidPublicKey)
	}

	serviceCert, err := keygen.DecodeCertificate(req.Service.Certificate)
	if err != nil {
		return exit(codes.InvalidArgument, "invalid service cert bytes", err)
	}

	if !a.cert.Equal(serviceCert) {
		return exit(codes.InvalidArgument, "mismatching service cert", ErrInvalidCertificate)
	}

	pubKey, err := keygen.DecodePublic(req.Id.PublicKey)
	if err != nil {
		return exit(codes.InvalidArgument, "invalid public key PEM bytes", err)
	}

	cert, err := keygen.DecodeCertificate(req.Id.Certificate)
	if err != nil {
		return exit(codes.InvalidArgument, "invalid cert bytes", err)
	}

	storedPub, _, err := a.services.GetService(ctx, req.Name)
	if err != nil {
		return exit(codes.Internal, "failed to fetch stored public key", err)
	}

	pub, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return exit(codes.InvalidArgument, "failed to retrieve public key from certificate", ErrInvalidCertificate)
	}

	if !pub.Equal(storedPub) {
		return exit(codes.PermissionDenied, "mismatching public keys", ErrInvalidPublicKey)
	}

	if !pubKey.Equal(storedPub) {
		return exit(codes.PermissionDenied, "mismatching public keys", ErrInvalidPublicKey)
	}

	if time.Now().After(cert.NotAfter) {
		a.logger.DebugContext(ctx, "expired certificate",
			slog.Time("expiry", cert.NotAfter), slog.String("service", req.Name))

		return exit(codes.InvalidArgument, "expired certificate", ErrInvalidCertificate)
	}

	// TODO: challenge should expire
	challenge := a.random.Random()

	if err = a.challenges.CreateChallenge(ctx, req.Name, challenge); err != nil {
		return exit(codes.Internal, "failed to store challenge", err)
	}

	return &pb.LoginResponse{Challenge: challenge}, nil
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

	challenge, err := a.challenges.GetChallenge(ctx, req.Name)
	if err != nil {
		// TODO: handle if not exists
		return exit(codes.Internal, "failed to get challenge", err)
	}

	storedPub, _, err := a.services.GetService(ctx, req.Name)
	if err != nil {
		// TODO: handle if not exists
		return exit(codes.Internal, "failed to get service details", err)
	}

	pub, err := keygen.DecodePublic(storedPub)
	if err != nil {
		return exit(codes.Internal, "failed to decode public key", err)
	}

	h := sha256.Sum256(challenge)

	if !ecdsa.VerifyASN1(pub, h[:], req.SignedChallenge) {
		return exit(codes.InvalidArgument, "invalid signature", ErrInvalidSignature)
	}

	// TODO: generate token
	// TODO: store token + expiry
	// TODO: return token + expiry

	return nil, nil
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

	cert, err := keygen.DecodeCertificate(req.Certificate)
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
