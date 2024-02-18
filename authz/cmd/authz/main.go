package main

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zalgonoise/x/authz/keygen"
	"github.com/zalgonoise/x/authz/log"
	"github.com/zalgonoise/x/authz/metrics"
	pb "github.com/zalgonoise/x/authz/pb/authz/v1"
	"github.com/zalgonoise/x/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/zalgonoise/x/authz/ca"
	"github.com/zalgonoise/x/authz/database"
	"github.com/zalgonoise/x/authz/grpcserver"
	"github.com/zalgonoise/x/authz/httpserver"
	"github.com/zalgonoise/x/authz/repository"
)

var modes = []string{"ca"}

func main() {
	logger := log.New("debug")
	runner := cli.NewRunner("authz",
		cli.WithOneOf(modes...),
		cli.WithExecutors(map[string]cli.Executor{
			"ca": cli.Executable(ExecCertificateAuthority),
		}),
	)

	code, err := runner.Run(logger)
	if err != nil {
		logger.ErrorContext(context.Background(), "runtime error", slog.String("error", err.Error()))
	}

	os.Exit(code)
}

const (
	defaultLocal = "local"
	defaultDB    = "local/ca.db"
	defaultKey   = "local/key.pem"

	defaultHTTPPort = 8080
	defaultGRPCPort = 8081

	shutdownTimeout = time.Minute
)

var (
	ErrIsDir          = errors.New("path cannot be a directory")
	ErrNotDir         = errors.New("path must be a directory")
	ErrFailedCreateDB = errors.New("failed to create database file")
	ErrInvalidKey     = errors.New("invalid ECDSA private key")
)

func ExecCertificateAuthority(ctx context.Context, logger *slog.Logger, args []string) (int, error) {
	fs := flag.NewFlagSet("ca", flag.ExitOnError)

	// TODO: add configurable tracing backend
	dbURI := fs.String("db", "", "the path to the SQLite DB file to store services and their certificates")
	privateKey := fs.String("private-key", "", "the path to the ECDSA private key file to use for the certificate authority")
	httpPort := fs.Int("http-port", defaultHTTPPort, "the exposed HTTP port for the CA's API")
	grpcPort := fs.Int("grpc-port", defaultGRPCPort, "the exposed gRPC port for the CA's API")
	dur := fs.Int("dur", 24, "duration to use on new certificate's expiry")

	if *httpPort <= 0 {
		*httpPort = defaultHTTPPort
	}

	if *grpcPort <= 0 {
		*grpcPort = defaultGRPCPort
	}

	if *dbURI == "" {
		if err := checkOrCreateDir(defaultLocal); err != nil {
			return 1, err
		}
		*dbURI = defaultDB
	}

	logger.DebugContext(ctx, "preparing database")

	// TODO: add tracer init logic

	db, err := database.Open(*dbURI)
	if err != nil {
		return 1, err
	}

	defer db.Close()

	if err = database.Migrate(ctx, db, database.CAService, logger); err != nil {
		return 1, err
	}

	logger.DebugContext(ctx, "database is ready")
	logger.DebugContext(ctx, "preparing private key")

	if *privateKey == "" {
		if err = checkOrCreateDir(defaultLocal); err != nil {
			return 1, err
		}

		*privateKey = defaultKey
	}

	key, err := openOrCreateKey(*privateKey)
	if err != nil {
		return 1, err
	}

	logger.DebugContext(ctx, "private key is ready")
	logger.DebugContext(ctx, "preparing CA service")

	m := metrics.NewMetrics()

	// TODO: add configurable cleanup cron
	repo, err := repository.NewCertificateAuthority(db,
		repository.WithLogger(logger),
		repository.WithMetrics(m),
	)
	if err != nil {
		return 1, err
	}

	caService, err := ca.NewCertificateAuthority(
		repo,
		key,
		ca.WithLogger(logger),
		ca.WithTemplate(ca.WithDurMonth(*dur)),
	)

	logger.DebugContext(ctx, "CA service is ready")
	logger.DebugContext(ctx, "preparing HTTP server")

	httpServer, err := httpserver.NewServer(fmt.Sprintf(":%d", *httpPort))
	if err != nil {
		return 1, err
	}

	logger.DebugContext(ctx, "HTTP server is ready", slog.Int("port", *httpPort))
	logger.DebugContext(ctx, "preparing gRPC server")

	grpcServer, err := runGRPCServer(ctx, *grpcPort, caService, nil, httpServer, logger, m)
	if err != nil {
		return 1, err
	}

	logger.DebugContext(ctx, "gRPC server is ready", slog.Int("port", *grpcPort))
	logger.DebugContext(ctx, "setting up metrics handler")

	if err = registerMetrics(m, httpServer); err != nil {
		return 1, err
	}

	logger.DebugContext(ctx, "metrics handler is ready", slog.String("endpoint", "/metrics"))
	logger.DebugContext(ctx, "serving requests")

	go runHTTPServer(ctx, *httpPort, httpServer, logger)

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	<-signalChannel

	logger.InfoContext(ctx, "shutting down", slog.Duration("timeout", shutdownTimeout))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err = httpServer.Shutdown(shutdownCtx); err != nil {
		return 1, err
	}

	grpcServer.Shutdown(shutdownCtx)

	return 0, nil
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

func checkOrCreateDir(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err = os.Mkdir(defaultLocal, 0750); err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}

			return nil
		}

		return err
	}

	stat, err := f.Stat()
	if err != nil {
		return err
	}

	if !stat.IsDir() {
		return ErrNotDir
	}

	return nil
}

func openOrCreateKey(path string) (*ecdsa.PrivateKey, error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			f, err := os.Create(path)
			if err != nil {
				return nil, err
			}

			defer f.Close()

			key, err := keygen.New()
			if err != nil {
				return nil, err
			}

			keyPEM, err := keygen.EncodePrivate(key)
			if err != nil {
				return nil, err
			}

			_, err = f.Write(keyPEM)
			if err != nil {
				return nil, err
			}

			return key, nil
		}

		return nil, err
	}

	buf, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return keygen.DecodePrivate(buf)
}
