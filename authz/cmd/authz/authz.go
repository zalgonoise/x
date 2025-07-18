package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"go.opentelemetry.io/otel/trace/noop"

	"github.com/zalgonoise/x/authz/internal/authz"
	"github.com/zalgonoise/x/authz/internal/config"
	"github.com/zalgonoise/x/authz/internal/database"
	"github.com/zalgonoise/x/authz/internal/httpserver"
	"github.com/zalgonoise/x/authz/internal/metrics"
	"github.com/zalgonoise/x/authz/internal/randomizer"
	"github.com/zalgonoise/x/authz/internal/repository"
	"github.com/zalgonoise/x/authz/internal/tracing"
	pb "github.com/zalgonoise/x/authz/pb/authz/v1"
)

func ExecAuthz(ctx context.Context, logger *slog.Logger, args []string) (int, error) {
	conf, err := config.New(args)
	if err != nil {
		return 1, err
	}

	logger.DebugContext(ctx, "preparing tracer")

	var (
		tracer      = noop.NewTracerProvider().Tracer(tracing.ServiceNameAuthz)
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

		tracingDone, err = tracing.Init(ctx, tracing.ServiceNameAuthz, exporter)
		if err != nil {
			return 1, err
		}

		tracer = tracing.Tracer(tracing.ServiceNameAuthz)
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

	if err = database.Migrate(ctx, db, database.AuthzService, logger); err != nil {
		return 1, err
	}

	logger.DebugContext(ctx, "database is ready")
	logger.DebugContext(ctx, "preparing private key")

	if conf.PrivateKey == "" {
		if err = checkOrCreateDir(defaultLocal); err != nil {
			return 1, err
		}

		conf.PrivateKey = defaultPrivateKey
	}

	key, err := openOrCreateKey(conf.PrivateKey)
	if err != nil {
		return 1, err
	}

	logger.DebugContext(ctx, "private key is ready")
	logger.DebugContext(ctx, "preparing authz service")

	m := metrics.NewMetrics()

	servicesRepo, err := repository.NewService(db,
		repository.WithLogger(logger),
		repository.WithMetrics(m),
		repository.WithTrace(tracer),
		repository.WithCleanupTimeout(conf.Database.CleanupTimeout),
		repository.WithCleanupSchedule(conf.Database.CleanupSchedule),
	)
	if err != nil {
		return 1, err
	}

	tokensRepo, err := repository.NewToken(db,
		repository.WithLogger(logger),
		repository.WithMetrics(m),
		repository.WithTrace(tracer),
		repository.WithCleanupTimeout(conf.Database.CleanupTimeout),
		repository.WithCleanupSchedule(conf.Database.CleanupSchedule),
	)
	if err != nil {
		return 1, err
	}

	authzService, err := authz.NewAuthz(
		conf.Name, conf.Authz.CAURL,
		key,
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
		return 1, err
	}

	logger.DebugContext(ctx, "authz service is ready")
	logger.DebugContext(ctx, "preparing HTTP server")

	httpServer, err := httpserver.NewServer(fmt.Sprintf(":%d", conf.HTTPPort))
	if err != nil {
		return 1, err
	}

	logger.DebugContext(ctx, "HTTP server is ready", slog.Int("port", conf.HTTPPort))
	logger.DebugContext(ctx, "preparing gRPC server")

	grpcServer, err := runGRPCServer(ctx, conf.GRPCPort, authzService, authzService, httpServer, logger, m)
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

	if err = authzService.Shutdown(shutdownCtx); err != nil {
		return 1, err
	}

	return 0, nil
}
