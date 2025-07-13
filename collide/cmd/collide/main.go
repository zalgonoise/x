package main

import (
	"context"
	"errors"
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
	"go.uber.org/automaxprocs/maxprocs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/zalgonoise/x/cli/v2"

	"github.com/zalgonoise/x/collide/internal/config"
	"github.com/zalgonoise/x/collide/internal/grpcserver"
	"github.com/zalgonoise/x/collide/internal/httpserver"
	"github.com/zalgonoise/x/collide/internal/log"
	"github.com/zalgonoise/x/collide/internal/metrics"
	"github.com/zalgonoise/x/collide/internal/profiling"
	"github.com/zalgonoise/x/collide/internal/repository/memory"
	"github.com/zalgonoise/x/collide/internal/service"
	"github.com/zalgonoise/x/collide/internal/tracing"
	pb "github.com/zalgonoise/x/collide/pkg/api/pb/collide/v1"
)

type Metrics interface {
	IncListDistricts(ctx context.Context)
	IncListDistrictsFailed(ctx context.Context)
	ObserveListDistrictsLatency(ctx context.Context, duration time.Duration)
	IncListAllTracksByDistrict(ctx context.Context, district string)
	IncListAllTracksByDistrictFailed(ctx context.Context, district string)
	ObserveListAllTracksByDistrictLatency(ctx context.Context, duration time.Duration, district string)
	IncListDriftTracksByDistrict(ctx context.Context, district string)
	IncListDriftTracksByDistrictFailed(ctx context.Context, district string)
	ObserveListDriftTracksByDistrictLatency(ctx context.Context, duration time.Duration, district string)
	IncGetAlternativesByDistrictAndTrack(ctx context.Context, district string, track string)
	IncGetAlternativesByDistrictAndTrackFailed(ctx context.Context, district string, track string)
	ObserveGetAlternativesByDistrictAndTrackLatency(ctx context.Context, duration time.Duration, district string, track string)
	IncGetCollisionsByDistrictAndTrack(ctx context.Context, district string, track string)
	IncGetCollisionsByDistrictAndTrackFailed(ctx context.Context, district string, track string)
	ObserveGetCollisionsByDistrictAndTrackLatency(ctx context.Context, duration time.Duration, district string, track string)
}

var ErrInvalidMetricsType = errors.New("invalid metrics type")

func main() {
	logger := log.New("debug", true, true)

	runner := cli.NewRunner("collide",
		cli.WithExecutors(map[string]cli.Executor{
			"serve": cli.Executable(ExecServe),
		}),
	)

	code, err := runner.Run(logger)
	if err != nil {
		logger.ErrorContext(context.Background(), "runtime error", slog.String("error", err.Error()))
	}

	os.Exit(code)
}

func ExecServe(ctx context.Context, logger *slog.Logger, _ []string) (int, error) {
	// init config
	logger.InfoContext(ctx, "loading config")
	cfg, err := config.New()
	if err != nil {
		return 1, err
	}

	_, err = maxprocs.Set(maxprocs.Logger(func(s string, i ...interface{}) {
		logger.InfoContext(ctx, fmt.Sprintf(s, i...))
	}))
	if err != nil {
		return 1, err
	}

	// init tracing
	traceExporter, err := tracing.GRPCExporter(cfg.Tracing)
	if err != nil {
		logger.WarnContext(ctx, "defaulting to using a no-op trace exporter", slog.String("error", err.Error()))

		traceExporter = tracing.NoopExporter()
	}

	tracerDone, err := tracing.Init(ctx, traceExporter)
	if err != nil {
		return 1, err
	}

	tracer := tracing.Tracer("collide")

	// init profiling
	if cfg.Profiling.Enabled {
		profiler, err := profiling.New(cfg.Profiling.Name, cfg.Profiling.URI, cfg.Profiling.Tags, logger)
		switch {
		case err != nil:
			logger.WarnContext(ctx, "starting profiler",
				slog.String("error", err.Error()),
				slog.String("profiler_uri", cfg.Profiling.URI))

		default:
			defer func() {
				if err := profiler.Stop(); err != nil {
					logger.ErrorContext(ctx, "stopping profiler", slog.String("error", err.Error()))
				}
			}()
		}
	}

	// init HTTP server
	httpServer, err := httpserver.NewServer(fmt.Sprintf(":%d", cfg.HTTP.Port))
	if err != nil {
		return 1, err
	}

	// init metrics
	var m Metrics

	switch cfg.Metrics.URI {
	case "":
		m = metrics.NewPrometheus()

		if err := registerMetrics(httpServer, m); err != nil {
			return 1, err
		}
	default:
		metricsDone, err := metrics.Init(ctx, cfg.Metrics.URI)
		if err != nil {
			return 1, err
		}

		defer func() {
			if err := metricsDone(ctx); err != nil {
				logger.ErrorContext(ctx, "closing metrics exporter", slog.String("error", err.Error()))
			}
		}()

		if m, err = metrics.NewOtel(); err != nil {
			return 1, err
		}
	}

	// setup service
	f, err := os.Open(cfg.Tracks.Path)
	if err != nil {
		return 1, err
	}

	buf, err := io.ReadAll(f)
	if err != nil {
		return 1, errors.Join(err, f.Close())
	}

	if err := f.Close(); err != nil {
		return 1, err
	}

	repo := memory.New(logger, tracer)

	if err := repo.FromBytes(buf); err != nil {
		return 1, err
	}

	collideService := service.New(repo, m, logger, tracer)

	// init gRPC server
	grpcServer, err := initGRPCServer(ctx, cfg.HTTP.GRPCPort, collideService, httpServer, logger, m)
	if err != nil {
		return 1, err
	}

	go runHTTPServer(ctx, cfg.HTTP.Port, httpServer, logger)

	// handle graceful shutdown on exit signal
	return handleGracefulShutdown(ctx, logger, httpServer, grpcServer, tracerDone, []context.CancelFunc{})
}

func initGRPCServer(
	ctx context.Context,
	port int,
	collideService pb.CollideServiceServer,
	httpServer *httpserver.Server,
	logger *slog.Logger,
	m Metrics,
) (*grpcserver.Server, error) {
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	grpcServer := grpcserver.NewServer(
		[]grpc.UnaryServerInterceptor{
			logging.UnaryServerInterceptor(log.InterceptorLogger(logger), loggingOpts...),
		},
		[]grpc.StreamServerInterceptor{
			logging.StreamServerInterceptor(log.InterceptorLogger(logger), loggingOpts...),
		},
		m,
	)

	grpcServer.RegisterCollideServer(collideService)

	logger.InfoContext(ctx, "starting grpc server", slog.Int("port", port))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.ErrorContext(ctx, "grpc server error", slog.String("error", err.Error()))

			os.Exit(1)
		}
	}()

	grpcClient, err := grpc.NewClient(
		fmt.Sprintf("localhost:%d", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	serviceClient := pb.NewCollideServiceClient(grpcClient)

	if err = httpServer.RegisterCollideService(ctx, serviceClient); err != nil {
		return nil, err
	}

	return grpcServer, nil
}

func registerMetrics(
	httpServer *httpserver.Server,
	m Metrics,
) error {
	promMetrics, ok := m.(*metrics.Prometheus)
	if !ok {
		return ErrInvalidMetricsType
	}

	promRegistry, err := metrics.Registry(promMetrics)
	if err != nil {
		return err
	}

	return httpServer.RegisterHTTP(http.MethodGet, "/metrics",
		promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{Registry: promRegistry}))
}

func runHTTPServer(
	ctx context.Context,
	port int,
	server *httpserver.Server,
	logger *slog.Logger,
) {
	logger.InfoContext(ctx, "starting http server", slog.Int("port", port))

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.ErrorContext(ctx, "http server error", slog.String("error", err.Error()))

		os.Exit(1)
	}
}

func handleGracefulShutdown(
	ctx context.Context,
	logger *slog.Logger,
	httpServer *httpserver.Server,
	grpcServer *grpcserver.Server,
	tracerDone tracing.ShutdownFunc,
	cancelFuncs []context.CancelFunc,
) (int, error) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	<-signalChannel

	shutdownTimeout := 30 * time.Second

	logger.InfoContext(ctx, "shutting down", slog.Duration("timeout", shutdownTimeout))

	for i := range cancelFuncs {
		cancelFuncs[i]()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return 1, err
	}

	grpcServer.Shutdown(shutdownCtx)

	if err := tracerDone(shutdownCtx); err != nil {
		return 1, err
	}

	return 0, nil
}
