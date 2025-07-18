package main

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/zalgonoise/x/cli/v2"

	"github.com/zalgonoise/x/authz/internal/grpcserver"
	"github.com/zalgonoise/x/authz/internal/httpserver"
	"github.com/zalgonoise/x/authz/internal/keygen"
	"github.com/zalgonoise/x/authz/internal/log"
	"github.com/zalgonoise/x/authz/internal/metrics"
	pb "github.com/zalgonoise/x/authz/pb/authz/v1"
)

func main() {
	logger := log.New("debug")
	runner := cli.NewRunner("authz",
		cli.WithExecutors(map[string]cli.Executor{
			"ca":    cli.Executable(ExecCertificateAuthority),
			"authz": cli.Executable(ExecAuthz),
		}),
	)

	cli.Run(runner, logger)
}

const (
	defaultLocal      = "local"
	defaultDB         = "local/ca.db"
	defaultPrivateKey = "local/key.pem"
	defaultPublicKey  = "local/pub.pem"

	shutdownTimeout = time.Minute
)

var (
	ErrNotDir = errors.New("path must be a directory")
)

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
			keyFile, err := os.Create(path)
			if err != nil {
				return nil, err
			}

			key, err := keygen.New()
			if err != nil {
				return nil, err
			}

			if err := writePrivate(key, keyFile); err != nil {
				return nil, err
			}

			// also create a public key
			pubPath := defaultPublicKey
			if path != defaultPrivateKey {
				pubPath = path + ".pub"
			}

			pubFile, err := os.Create(pubPath)
			if err != nil {
				return nil, err
			}

			if err := writePublic(&key.PublicKey, pubFile); err != nil {
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

func writePrivate(key *ecdsa.PrivateKey, w io.WriteCloser) error {
	defer w.Close()

	keyPEM, err := keygen.EncodePrivate(key)
	if err != nil {
		return err
	}

	if _, err := w.Write(keyPEM); err != nil {
		return err
	}

	return nil
}

func writePublic(key *ecdsa.PublicKey, w io.WriteCloser) error {
	defer w.Close()

	keyPEM, err := keygen.EncodePublic(key)
	if err != nil {
		return err
	}

	if _, err := w.Write(keyPEM); err != nil {
		return err
	}

	return nil
}
