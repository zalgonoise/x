//go:build integration

package authz_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/require"
	"github.com/zalgonoise/x/authz/authz"
	"github.com/zalgonoise/x/authz/ca"
	"github.com/zalgonoise/x/authz/certs"
	"github.com/zalgonoise/x/authz/config"
	"github.com/zalgonoise/x/authz/database"
	"github.com/zalgonoise/x/authz/grpcserver"
	"github.com/zalgonoise/x/authz/httpserver"
	"github.com/zalgonoise/x/authz/keygen"
	"github.com/zalgonoise/x/authz/log"
	"github.com/zalgonoise/x/authz/metrics"
	pb "github.com/zalgonoise/x/authz/pb/authz/v1"
	"github.com/zalgonoise/x/authz/randomizer"
	"github.com/zalgonoise/x/authz/repository"
	"github.com/zalgonoise/x/authz/tracing"
	"go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestAuthz(t *testing.T) {
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := log.New("debug")

	service, done, errs := initServices(ctx, logger)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case err := <-errs:
				t.Error(err)
				t.Fail()
			}
		}
	}()

	defer func() {
		if err := done(); err != nil {
			t.Error(err)
			t.Fail()
		}
	}()

	// sample keys
	key, err := keygen.New()
	require.NoError(t, err)

	privPEM, err := keygen.EncodePrivate(key)
	require.NoError(t, err)

	pubPEM, err := keygen.EncodePublic(&key.PublicKey)
	require.NoError(t, err)

	t.Run("SignUp", func(t *testing.T) {
		for _, testcase := range []struct {
			name  string
			req   *pb.SignUpRequest
			fails bool
		}{
			{
				name: "Success/Simple",
				req: &pb.SignUpRequest{
					Name:      "test.simple",
					PublicKey: pubPEM,
				},
			},
			{
				name: "Success/FetchAgain",
				req: &pb.SignUpRequest{
					Name:      "test.simple",
					PublicKey: pubPEM,
				},
			},
			{
				name: "Fail/PrivateInsteadOfPublic",
				req: &pb.SignUpRequest{
					Name:      "test.fail.priv-key",
					PublicKey: privPEM,
				},
				fails: true,
			},
		} {
			t.Run(testcase.name, func(t *testing.T) {
				ctx := context.Background()

				res, err := service.SignUp(ctx, testcase.req)
				if err != nil {
					require.True(t, testcase.fails)

					return
				}

				require.NotNil(t, res.Certificate)
				require.NotNil(t, res.Service.PublicKey)
				require.NotNil(t, res.Service.Certificate)
			})
		}
	})
	t.Run("Login", func(t *testing.T) {})
	t.Run("Token", func(t *testing.T) {})
	t.Run("VerifyToken", func(t *testing.T) {})
	t.Run("Register", func(t *testing.T) {})
	t.Run("GetCertificate", func(t *testing.T) {})
	t.Run("VerifyCertificate", func(t *testing.T) {})
	t.Run("DeleteService", func(t *testing.T) {})
	t.Run("PublicKey", func(t *testing.T) {})
}

func cleanup() error {
	if err := os.RemoveAll("./testdata"); err != nil {
		return err
	}

	return os.Mkdir("./testdata", 0755)
}

func initServices(ctx context.Context, logger *slog.Logger) (*authz.Authz, func() error, <-chan error) {
	errs := make(chan error)

	go func() {
		if err := initCA(ctx, 8080, 8081, logger); err != nil {
			errs <- err
		}
	}()

	service, done, err := initAuthz(ctx, "service", "localhost:8081", 8082, 8083, logger)
	if err != nil {
		errs <- err

		return nil, nil, errs
	}

	return service, done, errs
}

func initCA(
	ctx context.Context,
	httpPort, grpcPort int,
	logger *slog.Logger,
) error {
	_, privateKey, err := newKeys("ca")
	if err != nil {
		return err
	}

	conf := config.Config{
		HTTPPort: httpPort,
		GRPCPort: grpcPort,
		Name:     "authz.test.ca",
		CA: config.CA{
			CertDurMonths: 12,
		},
		Database: config.Database{
			URI:             "./testdata/ca.db",
			CleanupTimeout:  5 * time.Minute,
			CleanupSchedule: "0 6 * * *",
		},
	}

	tracer := noop.NewTracerProvider().Tracer(tracing.ServiceNameCA)
	tracingDone := func(context.Context) error { return nil }

	db, err := database.Open(conf.Database.URI)
	if err != nil {
		return err
	}

	if err = database.Migrate(ctx, db, database.CAService, logger); err != nil {
		return err
	}

	m := metrics.NewMetrics()

	repo, err := repository.NewService(db,
		repository.WithLogger(logger),
		repository.WithMetrics(m),
		repository.WithTrace(tracer),
		repository.WithCleanupTimeout(conf.Database.CleanupTimeout),
		repository.WithCleanupSchedule(conf.Database.CleanupSchedule),
	)

	caService, err := ca.NewCertificateAuthority(
		repo,
		privateKey,
		ca.WithLogger(logger),
		ca.WithMetrics(m),
		ca.WithTracer(tracer),
		ca.WithTemplate(
			certs.WithName(pkix.Name{CommonName: conf.Name}),
			certs.WithDurMonth(conf.CA.CertDurMonths),
		),
	)

	httpServer, err := httpserver.NewServer(fmt.Sprintf(":%d", conf.HTTPPort))
	if err != nil {
		return err
	}

	grpcServer, err := runGRPCServer(ctx, conf.GRPCPort, caService, nil, httpServer, logger, m)
	if err != nil {
		return err
	}

	if err = registerMetrics(m, httpServer); err != nil {
		return err
	}

	go runHTTPServer(ctx, conf.HTTPPort, httpServer, logger)

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	done := func() error {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		if err = httpServer.Shutdown(shutdownCtx); err != nil {
			return err
		}

		grpcServer.Shutdown(shutdownCtx)

		if err = tracingDone(shutdownCtx); err != nil {
			return err
		}

		if err = repo.Close(); err != nil {
			return err
		}

		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return done()
		case sig := <-signalChannel:
			errSignal := fmt.Errorf("received exit signal: %s", sig.String())

			return errors.Join(errSignal, done())
		}
	}

}

func initAuthz(
	ctx context.Context,
	name, caAddress string,
	httpPort, grpcPort int,
	logger *slog.Logger,
) (*authz.Authz, func() error, error) {
	_, privateKey, err := newKeys(name)
	if err != nil {
		return nil, nil, err
	}

	conf := config.Config{
		HTTPPort: httpPort,
		GRPCPort: grpcPort,
		Name:     fmt.Sprintf("authz.test.%s", name),
		Authz: config.Authz{
			CAURL:         caAddress,
			RandSize:      128,
			CertDurMonths: 12,
			ChallengeDur:  time.Minute * 10,
			TokenDur:      time.Hour,
		},
		Database: config.Database{
			URI:             "./testdata/authz.db",
			CleanupTimeout:  5 * time.Minute,
			CleanupSchedule: "0 6 * * *",
		},
	}

	tracer := noop.NewTracerProvider().Tracer(tracing.ServiceNameCA)
	tracingDone := func(context.Context) error { return nil }

	db, err := database.Open(conf.Database.URI)
	if err != nil {
		return nil, nil, err
	}

	if err = database.Migrate(ctx, db, database.AuthzService, logger); err != nil {
		return nil, nil, err
	}

	m := metrics.NewMetrics()

	servicesRepo, err := repository.NewService(db,
		repository.WithLogger(logger),
		repository.WithMetrics(m),
		repository.WithTrace(tracer),
		repository.WithCleanupTimeout(conf.Database.CleanupTimeout),
		repository.WithCleanupSchedule(conf.Database.CleanupSchedule),
	)
	if err != nil {
		return nil, nil, err
	}

	tokensRepo, err := repository.NewToken(db,
		repository.WithLogger(logger),
		repository.WithMetrics(m),
		repository.WithTrace(tracer),
		repository.WithCleanupTimeout(conf.Database.CleanupTimeout),
		repository.WithCleanupSchedule(conf.Database.CleanupSchedule),
	)
	if err != nil {
		return nil, nil, err
	}

	service, err := authz.NewAuthz(
		conf.Name, conf.Authz.CAURL,
		privateKey,
		servicesRepo, tokensRepo,
		randomizer.New(conf.Authz.RandSize),
		authz.WithDurMonth(conf.Authz.CertDurMonths),
		authz.WithChallengeExpiry(conf.Authz.ChallengeDur),
		authz.WithTokenExpiry(conf.Authz.TokenDur),
		authz.WithCSR(&pb.CSR{
			Subject: &pb.Subject{
				CommonName: conf.Name,
			},
		}),
		authz.WithLogger(logger),
		authz.WithMetrics(m),
		authz.WithTracer(tracer),
	)
	if err != nil {
		return nil, nil, err
	}

	return service, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		if err = tracingDone(ctx); err != nil {
			return err
		}

		if err = servicesRepo.Close(); err != nil {
			return err
		}

		if err = tokensRepo.Close(); err != nil {
			return err
		}
		return nil
	}, nil
}

func newKeys(name string) (string, *ecdsa.PrivateKey, error) {
	path := "./testdata/" + name + ".key.pem"

	keyF, err := os.Create(path)
	if err != nil {
		return "", nil, err
	}

	defer keyF.Close()

	key, err := keygen.New()

	keyPEM, err := keygen.EncodePrivate(key)
	if err != nil {
		return "", nil, err
	}

	if _, err := keyF.Write(keyPEM); err != nil {
		return "", nil, err
	}

	return path, key, nil
}

func runGRPCServer(
	ctx context.Context,
	port int,
	caService pb.CertificateAuthorityServer,
	authzService pb.AuthzServer,
	server *httpserver.Server,
	logger *slog.Logger,
	m *metrics.Metrics,
) (*grpcserver.Server, error) {
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	grpcServer := grpcserver.NewServer(m,
		[]grpc.UnaryServerInterceptor{
			logging.UnaryServerInterceptor(log.InterceptorLogger(logger), loggingOpts...),
		}, []grpc.StreamServerInterceptor{
			logging.StreamServerInterceptor(log.InterceptorLogger(logger), loggingOpts...),
		},
	)

	if caService != nil {
		grpcServer.RegisterCertificateAuthorityServer(caService)
	}

	if authzService != nil {
		grpcServer.RegisterAuthzServer(authzService)
	}

	logger.InfoContext(ctx, "listening on gRPC", slog.Int("port", port))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	go func() {
		if serveErr := grpcServer.Serve(lis); serveErr != nil {
			logger.ErrorContext(ctx, "failed to start gRPC server",
				slog.String("error", serveErr.Error()),
				slog.Int("port", port),
			)
			os.Exit(1)
		}
	}()

	grpcClient, err := grpc.Dial(
		fmt.Sprintf("localhost:%d", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	if caService != nil {
		caServiceClient := pb.NewCertificateAuthorityClient(grpcClient)

		if err = server.RegisterCA(ctx, caServiceClient); err != nil {
			return nil, err
		}
	}

	if authzService != nil {
		authzServiceClient := pb.NewAuthzClient(grpcClient)

		if err = server.RegisterAuthz(ctx, authzServiceClient); err != nil {
			return nil, err
		}
	}

	return grpcServer, nil
}

func registerMetrics(m *metrics.Metrics, httpServer *httpserver.Server) error {
	reg, err := m.Registry()
	if err != nil {
		return err
	}

	return httpServer.RegisterHTTP(http.MethodGet, "/metrics",
		promhttp.HandlerFor(reg,
			promhttp.HandlerOpts{Registry: reg}))
}

func runHTTPServer(ctx context.Context, port int, httpServer *httpserver.Server, logger *slog.Logger) {
	logger.InfoContext(ctx, "listening on http", slog.Int("port", port))

	err := httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.ErrorContext(ctx, "http server listen error", slog.String("error", err.Error()))
	}
}