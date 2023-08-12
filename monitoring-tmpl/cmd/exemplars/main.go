package main

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/zalgonoise/x/monitoring-tmpl/config"
	"github.com/zalgonoise/x/monitoring-tmpl/log"
	"github.com/zalgonoise/x/monitoring-tmpl/metrics"
	"github.com/zalgonoise/x/monitoring-tmpl/service"
	"github.com/zalgonoise/x/monitoring-tmpl/tracing"
)

func main() {
	err, exitCode := run()

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "fatal: %s\n", err)
	}

	os.Exit(exitCode)
}

func run() (error, int) {
	ctx := context.Background()
	cfg, err := config.NewServiceConfig()
	if err != nil {
		return err, 1
	}

	logger := slog.New(log.NewSpanContextHandler(
		log.WithHandler(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			AddSource: true,
		})),
		log.WithSpanID(),
	))

	exporter, err := tracing.GRPCExporter(ctx, cfg.TracerURI)
	if err != nil {
		return err, 1
	}

	tracerDone, err := tracing.Init(ctx, exporter)
	if err != nil {
		return err, 1
	}

	defer tracerDone(ctx)

	m := metrics.NewMetrics()

	reg, err := m.Registry()
	if err != nil {
		return err, 1
	}

	metricsServer := metrics.NewServer(cfg.MetricsPort, reg)
	defer metricsServer.Shutdown(ctx)

	var handler service.Service = service.NewHandler(cfg.Threshold)
	handler = service.WithLogs(handler, logger)
	handler = service.WithMetrics(handler, m)
	handler = service.WithTrace(handler, tracing.Tracer())

	rng := newRNG()

	ctx, done := context.WithTimeout(ctx, cfg.Duration)
	defer done()

	for {
		select {
		case <-ctx.Done():
			return nil, 0
		default:
			_ = handler.Handle(
				context.Background(),
				rng.Intn(cfg.MaxInputValue),
			)
		}
	}
}

func newRNG() *rand.Rand {
	return rand.New(rand.NewSource(
		int64(float64(time.Now().Unix()) / math.Pi),
	))
}
