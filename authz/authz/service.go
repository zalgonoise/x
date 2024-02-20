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
	"github.com/zalgonoise/x/authz/keygen"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/zalgonoise/x/authz/log"
	"github.com/zalgonoise/x/authz/metrics"
	pb "github.com/zalgonoise/x/authz/pb/authz/v1"
)

const defaultConnTimeout = time.Minute

type Repository interface {
	Get(ctx context.Context, service string) (pubKey []byte, cert []byte, err error)
	Create(ctx context.Context, service string, pubKey []byte, cert []byte, expiry time.Time) (err error)
	Delete(ctx context.Context, service string) error
}

type Randomizer interface {
	Random() int64
}

type Metrics interface {
	RegisterCollector(collector prometheus.Collector)
}

type Authz struct {
	pb.UnimplementedAuthzServer
	caClient pb.CertificateAuthorityClient

	privateKey *ecdsa.PrivateKey
	cert       *x509.Certificate

	repo   Repository
	random Randomizer

	m      Metrics
	logger *slog.Logger
	tracer trace.Tracer
}

func NewAuthz(
	name, caAddress string,
	privateKey *ecdsa.PrivateKey,
	repo Repository,
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
		repo:       repo,
		random:     randomizer,
		m:          config.m,
		logger:     config.logger,
		tracer:     config.tracer,
	}, nil
}

func (a *Authz) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	return nil, nil
}

func (a *Authz) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	return nil, nil
}

func (a *Authz) Token(ctx context.Context, req *pb.TokenRequest) (*pb.TokenResponse, error) {
	return nil, nil
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
