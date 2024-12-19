package main

import (
	"context"
	"errors"
	"fmt"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/audio/sdk/audio/exporters/database"

	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/fft"
	"github.com/zalgonoise/x/audio/sdk/audio"
	"github.com/zalgonoise/x/audio/sdk/audio/compactors"
	"github.com/zalgonoise/x/audio/sdk/audio/consumers/httpaudio"
	"github.com/zalgonoise/x/audio/sdk/audio/exporters"
	"github.com/zalgonoise/x/audio/sdk/audio/exporters/stats/prom"
	"github.com/zalgonoise/x/audio/sdk/audio/exporters/stats/stdout"
	"github.com/zalgonoise/x/audio/sdk/audio/processors"
	"github.com/zalgonoise/x/audio/sdk/audio/registries/batchreg"
)

const (
	shutdownTimeout = 15 * time.Second
)

const (
	undefined = iota
	noHost
	withHost
)

func main() {
	code, err := run()
	if err != nil {
		slog.Error(
			"audio: runtime error",
			slog.String("error", err.Error()),
		)
	}

	os.Exit(code)
}

func run() (int, error) {
	logHandler := slog.NewTextHandler(os.Stderr, nil)
	logger := slog.New(logHandler)
	ctx := context.Background()

	//TODO: replace with an actual tracer implementation
	tracer := noop.NewTracerProvider().Tracer("noop")

	exporterMetrics := audio.NoOpExporterMetrics{}
	processorMetrics := audio.NoOpProcessorMetrics{}

	config, err := NewConfig()
	if err != nil {
		return 1, err
	}

	logger.InfoContext(ctx, "setting up consumer")

	consumer, err := httpaudio.New(
		httpaudio.WithTarget(config.Input),
		httpaudio.WithTimeout(config.Duration),
	)
	if err != nil {
		return 1, err
	}

	logger.InfoContext(ctx, "setting up exporters")

	var exp = make([]audio.Exporter, 0, 2)

	stats, err := newStatsExporter(ctx, config, logger, exporterMetrics, tracer)
	if err != nil {
		return 1, err
	}

	exp = append(exp, stats)

	if config.StorageURI != "" {
		logger.InfoContext(ctx, "setting up SQLite data hook")

		dataExporter, err := newDataExporter(ctx, config, logHandler)
		if err != nil {
			return 1, err
		}

		exp = append(exp, dataExporter)
	}

	logger.InfoContext(ctx, "setting up processor")

	proc := processors.PCM(exp, []cfg.Option[wav.Config]{
		wav.WithSize(config.BufferSize),
		wav.WithDuration(config.BufferDur),
		wav.WithRatio(config.BufferRatio),
	}, logger, processorMetrics, tracer,
	)

	ctx, cancel := context.WithTimeout(context.Background(), config.Duration)
	defer cancel()

	logger.InfoContext(ctx, "reading from consumer")

	reader, err := consumer.Consume(ctx)
	if err != nil {
		return 1, err
	}

	logger.InfoContext(ctx, "processing signal")

	go proc.Process(ctx, reader)

	errs := proc.Err()

	defer func() {
		logger.InfoContext(ctx, "shutting down", slog.Duration("timeout", shutdownTimeout))

		if shutdownErr := audio.Shutdown(ctx, shutdownTimeout, consumer, proc); shutdownErr != nil {
			logger.WarnContext(ctx, "error when shutting down", slog.String("error", shutdownErr.Error()))
		}
	}()

	for {
		select {
		case <-ctx.Done():
			logger.InfoContext(ctx, "exiting")

			return 0, nil
		case procErr, ok := <-errs:
			if !ok || procErr == nil || errors.Is(procErr, audio.ErrHaltSignal) {
				return 0, nil
			}

			return 1, procErr
		}
	}
}

func newDataExporter(ctx context.Context, config *Config, logHandler slog.Handler) (audio.Exporter, error) {
	logger := slog.New(logHandler)
	db, err := database.OpenSQLite(config.StorageURI, database.ReadWritePragmas(), logger)
	if err != nil {
		return nil, err
	}

	if err := database.MigrateSQLite(ctx, db); err != nil {
		return nil, err
	}

	return exporters.NewSQLiteExporter(database.NewRepository(db), exporters.WithFlushDuration(config.BufferDur))
}

func newStatsExporter(
	ctx context.Context, config *Config,
	logger *slog.Logger, metrics audio.ExporterMetrics, tracer trace.Tracer,
) (audio.Exporter, error) {
	var (
		exporterOpts = newExporterOpts(config, logger.Handler())
		exporter     audio.Exporter
		err          error
	)

	switch config.OutputType {
	case "prom", "prometheus":
		port, err := getPort(config.Output)
		if err != nil {
			logger.WarnContext(ctx, "failed to parse port; using defaults",
				slog.String("input_value", config.Output),
				slog.String("error", err.Error()),
				slog.Int("default_port", port),
			)
		}

		exporter, err = prom.ToProm(port, exporterOpts, logger, tracer)
		if err != nil {
			return audio.NoOpExporter(), err
		}
	default:
		exporter, err = stdout.ToLogger(exporterOpts, logger, metrics, tracer)
		if err != nil {
			return audio.NoOpExporter(), err
		}
	}

	return exporter, nil
}

func newExporterOpts(config *Config, logHandler slog.Handler) []cfg.Option[*exporters.StatsConfig] {
	return []cfg.Option[*exporters.StatsConfig]{
		exporters.WithLogHandler(logHandler),
		newPeaksOpt(config),
		newSpectrumOpt(config),
	}
}

func newPeaksOpt(config *Config) cfg.Option[*exporters.StatsConfig] {
	if config.Mode == "spectrum" {
		return cfg.NoOp[*exporters.StatsConfig]{}
	}

	if config.Batch {
		return exporters.WithBatchedPeaks(
			batchreg.WithBatchSize[float64](config.BatchSize),
			batchreg.WithFlushFrequency[float64](config.BatchFrequency),
			batchreg.WithCompactor[float64](compactors.Max[float64]),
		)
	}

	return exporters.WithPeaks()
}

func newSpectrumOpt(config *Config) cfg.Option[*exporters.StatsConfig] {
	if config.Mode == "peaks" {
		return cfg.NoOp[*exporters.StatsConfig]{}
	}

	if config.Batch {
		return exporters.WithBatchedSpectrum(config.BucketSize,
			batchreg.WithBatchSize[[]fft.FrequencyPower](config.BatchSize),
			batchreg.WithFlushFrequency[[]fft.FrequencyPower](config.BatchFrequency),
			batchreg.WithCompactor[[]fft.FrequencyPower](compactors.UpperSpectra),
		)
	}

	return exporters.WithSpectrum(config.BucketSize)
}

func getPort(addr string) (port int, err error) {
	if addr == "" {
		return defaultPort, nil
	}

	split := strings.Split(addr, ":")

	switch len(split) {
	case undefined:
		return defaultPort, nil
	case noHost:
		port, err = strconv.Atoi(split[0])
		if err != nil {
			return defaultPort, err
		}

		return port, nil
	case withHost:
		port, err = strconv.Atoi(split[1])
		if err != nil {
			return defaultPort, err
		}

		return port, nil
	default:
		return defaultPort, fmt.Errorf("%w: %s", ErrInvalidPort, addr)
	}
}
