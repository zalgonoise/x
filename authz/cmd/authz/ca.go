package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/zalgonoise/x/authz/ca"
	"github.com/zalgonoise/x/authz/certs"
	"github.com/zalgonoise/x/authz/config"
	"github.com/zalgonoise/x/authz/database"
	"github.com/zalgonoise/x/authz/httpserver"
	"github.com/zalgonoise/x/authz/metrics"
	"github.com/zalgonoise/x/authz/repository"
	"github.com/zalgonoise/x/authz/tracing"
	"go.opentelemetry.io/otel/trace/noop"
)

func ExecCertificateAuthority(ctx context.Context, logger *slog.Logger, args []string) (int, error) {
	conf, err := config.New(args)
	if err != nil {
		return 1, err
	}

	logger.DebugContext(ctx, "preparing tracer")

	var (
		tracer      = noop.NewTracerProvider().Tracer(tracing.ServiceNameCA)
		tracingDone = func(context.Context) error { return nil }
	)

	if conf.Tracer.URI != "" {
		exporter, err := tracing.GRPCExporter(ctx, conf.Tracer.URI,
			tracing.WithCredentials(conf.Tracer.Username, conf.Tracer.Password),
			tracing.WithTimeout(conf.Tracer.ConnTimeout),
		)

		if err != nil {
			return 1, err
		}

		tracingDone, err = tracing.Init(ctx, tracing.ServiceNameCA, exporter)
		if err != nil {
			return 1, err
		}

		tracer = tracing.Tracer(tracing.ServiceNameCA)
	}

	logger.DebugContext(ctx, "preparing database")

	if conf.Database.URI == "" {
		if err := checkOrCreateDir(defaultLocal); err != nil {
			return 1, err
		}
		conf.Database.URI = defaultDB
	}

	db, err := database.Open(conf.Database.URI)
	if err != nil {
		return 1, err
	}

	if err = database.Migrate(ctx, db, database.CAService, logger); err != nil {
		return 1, err
	}

	logger.DebugContext(ctx, "database is ready")
	logger.DebugContext(ctx, "preparing private key")

	if conf.PrivateKey == "" {
		if err = checkOrCreateDir(defaultLocal); err != nil {
			return 1, err
		}

		conf.PrivateKey = defaultKey
	}

	key, err := openOrCreateKey(conf.PrivateKey)
	if err != nil {
		return 1, err
	}

	logger.DebugContext(ctx, "private key is ready")
	logger.DebugContext(ctx, "preparing CA service")

	m := metrics.NewMetrics()

	repo, err := repository.NewService(db,
		repository.WithLogger(logger),
		repository.WithMetrics(m),
		repository.WithTrace(tracer),
		repository.WithCleanupTimeout(conf.Database.CleanupTimeout),
		repository.WithCleanupSchedule(conf.Database.CleanupSchedule),
	)
	if err != nil {
		return 1, err
	}

	caService, err := ca.NewCertificateAuthority(
		repo,
		key,
		ca.WithLogger(logger),
		ca.WithMetrics(m),
		ca.WithTracer(tracer),
		ca.WithTemplate(certs.WithDurMonth(conf.CA.CertDurMonths)),
	)

	logger.DebugContext(ctx, "CA service is ready")
	logger.DebugContext(ctx, "preparing HTTP server")

	httpServer, err := httpserver.NewServer(fmt.Sprintf(":%d", conf.HTTPPort))
	if err != nil {
		return 1, err
	}

	logger.DebugContext(ctx, "HTTP server is ready", slog.Int("port", conf.HTTPPort))
	logger.DebugContext(ctx, "preparing gRPC server")

	grpcServer, err := runGRPCServer(ctx, conf.GRPCPort, caService, nil, httpServer, logger, m)
	if err != nil {
		return 1, err
	}

	logger.DebugContext(ctx, "gRPC server is ready", slog.Int("port", conf.GRPCPort))
	logger.DebugContext(ctx, "setting up metrics handler")

	if err = registerMetrics(m, httpServer); err != nil {
		return 1, err
	}

	logger.DebugContext(ctx, "metrics handler is ready", slog.String("endpoint", "/metrics"))
	logger.DebugContext(ctx, "serving requests")

	go runHTTPServer(ctx, conf.HTTPPort, httpServer, logger)

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

	if err = tracingDone(shutdownCtx); err != nil {
		return 1, err
	}

	if err = repo.Close(); err != nil {
		return 1, err
	}

	return 0, nil
}
