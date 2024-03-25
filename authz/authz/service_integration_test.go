//go:build integration

package authz_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"slices"
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

var caHTTPPort string
var caGRPCPort string

func TestMain(m *testing.M) {
	exitCode := func() int {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		logger := log.New("debug")
		errs := make(chan error)

		go initCA(ctx, 8080, 8081, logger, errs)

		go func() {
			for {
				select {
				case err, ok := <-errs:
					if ok && err != nil {
						logger.Error("init error", slog.String("error", err.Error()))

						os.Exit(1)
					}
				case <-ctx.Done():
					return
				}
			}
		}()

		return m.Run()
	}()

	os.Exit(exitCode)
}

func TestAuthz(t *testing.T) {
	// cleanup before and after running the tests
	require.NoError(t, cleanup())
	defer func() {
		require.NoError(t, cleanup())
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	logger := log.New("debug")

	service, done, err := initAuthz(ctx, "service", "127.0.0.1:8081", 8082, 8083, logger)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, done())
	}()

	//service, done, errs := initServices(ctx, logger)
	//go func() {
	//	for {
	//		select {
	//		case <-ctx.Done():
	//			return
	//		case err := <-errs:
	//			t.Error(err)
	//			t.Fail()
	//		}
	//	}
	//}()

	//defer func() {
	//	if err := done(); err != nil {
	//		t.Error(err)
	//		t.Fail()
	//	}
	//}()

	// sample keys
	key, err := keygen.New()
	require.NoError(t, err)

	privPEM, err := keygen.EncodePrivate(key)
	require.NoError(t, err)

	pubPEM, err := keygen.EncodePublic(&key.PublicKey)
	require.NoError(t, err)

	t.Run("SignUp", func(t *testing.T) {
		for _, testcase := range []struct {
			name    string
			service string
			pubKey  []byte
			fails   bool
		}{
			{
				name:    "Success/Simple",
				service: "test.signup.simple",
				pubKey:  pubPEM,
			},
			{
				name:    "Success/SignupAgain",
				service: "test.signup.simple",
				pubKey:  pubPEM,
			},
			{
				name:    "Fail/PrivateInsteadOfPublic",
				service: "test.signup.fail.priv-key",
				pubKey:  privPEM,
				fails:   true,
			},
		} {
			t.Run(testcase.name, func(t *testing.T) {
				ctx := context.Background()

				res, err := service.SignUp(ctx, &pb.SignUpRequest{
					Name:      testcase.service,
					PublicKey: testcase.pubKey,
				})
				if err != nil {
					require.True(t, testcase.fails)

					return
				}

				require.NotNil(t, res.Certificate)
				require.NotNil(t, res.ServiceCertificate)
			})
		}
	})

	t.Run("Client/RoundTrip", func(t *testing.T) {
		for _, testcase := range []struct {
			name    string
			service string
			pubKey  []byte
			privKey *ecdsa.PrivateKey
			fails   bool
		}{
			{
				name:    "Success/Simple",
				service: "test.login.simple",
				pubKey:  pubPEM,
				privKey: key,
			},
			{
				name:    "Success/Replay",
				service: "test.login.simple",
				pubKey:  pubPEM,
				privKey: key,
			},
		} {
			t.Run(testcase.name, func(t *testing.T) {
				ctx := context.Background()

				// signup
				res, err := service.SignUp(ctx, &pb.SignUpRequest{
					Name:      testcase.service,
					PublicKey: testcase.pubKey,
				})
				require.NoError(t, err)

				require.NotNil(t, res.Certificate)
				require.NotNil(t, res.ServiceCertificate)

				// login
				loginRes, err := service.Login(ctx, &pb.LoginRequest{
					IdCertificate:      res.Certificate,
					ServiceCertificate: res.ServiceCertificate,
				})
				require.NoError(t, err)
				require.NotNil(t, loginRes.Challenge)
				require.NotZero(t, loginRes.ExpiresOn)

				// get token
				signature, _, err := keygen.ECDSASigner{Priv: testcase.privKey}.Sign(loginRes.Challenge)
				require.NoError(t, err)

				tokenRes, err := service.Token(ctx, &pb.TokenRequest{
					Certificate:     res.Certificate,
					SignedChallenge: signature,
				})

				require.NoError(t, err)
				require.NotEmpty(t, tokenRes.Token)
				require.NotZero(t, tokenRes.ExpiresOn)

				// verify token
				_, err = service.VerifyToken(ctx, &pb.AuthRequest{
					Token: tokenRes.Token,
				})
				require.NoError(t, err)

				token, err := keygen.ParseToken([]byte(tokenRes.Token), nil)
				require.NoError(t, err)

				t.Logf("success:\nservice: %q;\nauthz-service: %q;\ntoken: %q\n",
					token.Claim.Service, token.Claim.Authz, tokenRes.Token,
				)
			})
		}
	})

	t.Run("AuthzAsCA/RoundTrip", func(t *testing.T) {
		for _, testcase := range []struct {
			name    string
			service string
			pubKey  []byte
		}{
			{
				name:    "Success/Simple",
				service: "test.registry.simple",
				pubKey:  pubPEM,
			},
		} {
			t.Run(testcase.name, func(t *testing.T) {
				ctx := context.Background()

				registerRes, err := service.Register(ctx, &pb.CertificateRequest{
					Service:   testcase.service,
					PublicKey: testcase.pubKey,
				})
				require.NoError(t, err)

				certRes, err := service.ListCertificates(ctx, &pb.CertificateRequest{
					Service:   testcase.service,
					PublicKey: testcase.pubKey,
				})
				require.NoError(t, err)

				require.True(t, slices.ContainsFunc(certRes.Certificates, func(response *pb.CertificateResponse) bool {
					return bytes.Equal(response.Certificate, registerRes.Certificate)
				}))

				_, err = service.VerifyCertificate(ctx, &pb.VerificationRequest{
					Service:     testcase.service,
					Certificate: registerRes.Certificate,
				})
				require.NoError(t, err)

				_, err = service.DeleteService(ctx, &pb.DeletionRequest{
					Service:   testcase.service,
					PublicKey: testcase.pubKey,
				})
				require.NoError(t, err)
			})
		}
	})

	t.Run("PublicKey", func(t *testing.T) {
		res, err := service.RootCertificate(ctx, &pb.RootCertificateRequest{})
		require.NoError(t, err)
		require.NotNil(t, res.Root)
	})
}

func cleanup() error {
	if err := os.Remove("./testdata/ca.db"); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.Remove("./testdata/authz.db"); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

func initServices(ctx context.Context, logger *slog.Logger) (*authz.Authz, func() error, <-chan error) {
	errs := make(chan error)
	//
	//caContainer, err := startCA()
	//if err != nil {
	//	errs <- err
	//
	//	return nil, nil, errs
	//}

	//go initCA(ctx, 8080, 8081, logger, errs)

	service, done, err := initAuthz(ctx, "service", "127.0.0.1:8081", 8082, 8083, logger)
	if err != nil {
		errs <- err

		return nil, nil, errs
	}

	return service, done, errs
}

func getKey(path string) (*ecdsa.PrivateKey, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	pem, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return keygen.DecodePrivate(pem)
}

func initCA(
	ctx context.Context,
	httpPort, grpcPort int,
	logger *slog.Logger,
	errs chan<- error,
) {
	m := metrics.NewMetrics()
	tracer := noop.NewTracerProvider().Tracer(tracing.ServiceNameCA)
	tracingDone := func(context.Context) error { return nil }

	closerFuncs := make([]func() error, 0, 5)

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

	privateKey, err := getKey("./testdata/ca.testkey_private.pem")
	if err != nil {
		errs <- err

		return
	}

	db, err := database.Open(conf.Database.URI)
	if err != nil {
		errs <- err

		return
	}
	closerFuncs = append(closerFuncs, db.Close)

	if err = database.Migrate(ctx, db, database.CAService, logger); err != nil {
		errs <- err

		for i := range closerFuncs {
			if err := closerFuncs[i](); err != nil {
				errs <- err
			}
		}

		return
	}

	repo, err := repository.NewService(db,
		repository.WithLogger(logger),
		repository.WithMetrics(m),
		repository.WithTrace(tracer),
		repository.WithCleanupTimeout(conf.Database.CleanupTimeout),
		repository.WithCleanupSchedule(conf.Database.CleanupSchedule),
	)
	if err != nil {
		errs <- err

		for i := range closerFuncs {
			if err := closerFuncs[i](); err != nil {
				errs <- err
			}
		}

		return
	}

	closerFuncs = append(closerFuncs, repo.Close)

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
	if err != nil {
		errs <- err

		for i := range closerFuncs {
			if err := closerFuncs[i](); err != nil {
				errs <- err
			}
		}

		return
	}

	httpServer, err := httpserver.NewServer(fmt.Sprintf(":%d", conf.HTTPPort))
	if err != nil {
		errs <- err

		for i := range closerFuncs {
			if err := closerFuncs[i](); err != nil {
				errs <- err
			}
		}

		return
	}

	closerFuncs = append(closerFuncs, func() error {
		return httpServer.Shutdown(context.Background())
	})

	grpcServer, err := runGRPCServer(ctx, conf.GRPCPort, caService, nil, httpServer, logger, m)
	if err != nil {
		errs <- err

		for i := range closerFuncs {
			if err := closerFuncs[i](); err != nil {
				errs <- err
			}
		}

		return
	}

	closerFuncs = append(closerFuncs, func() error {
		grpcServer.Shutdown(context.Background())

		return nil
	})

	if err = registerMetrics(m, httpServer); err != nil {
		errs <- err

		for i := range closerFuncs {
			if err := closerFuncs[i](); err != nil {
				errs <- err
			}
		}

		return
	}

	go runHTTPServer(ctx, conf.HTTPPort, httpServer, logger)

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	done := func() error {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		for i := range closerFuncs {
			if err := closerFuncs[i](); err != nil {
				errs <- err
			}
		}

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
			if err := done(); err != nil {
				errs <- err
			}

			return
		case sig := <-signalChannel:
			errSignal := fmt.Errorf("received exit signal: %s", sig.String())

			errs <- errors.Join(errSignal, done())

			return
		}
	}
}

func initAuthz(
	ctx context.Context,
	name, caAddress string,
	httpPort, grpcPort int,
	logger *slog.Logger,
) (*authz.Authz, func() error, error) {
	m := metrics.NewMetrics()
	tracer := noop.NewTracerProvider().Tracer(tracing.ServiceNameCA)
	tracingDone := func(context.Context) error { return nil }

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

	privateKey, err := getKey("./testdata/authz.testkey_private.pem")
	if err != nil {
		return nil, nil, err
	}

	db, err := database.Open(conf.Database.URI)
	if err != nil {
		return nil, nil, err
	}

	if err = database.Migrate(ctx, db, database.AuthzService, logger); err != nil {
		return nil, nil, err
	}

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
		fmt.Sprintf(":%d", port),
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
