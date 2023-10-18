package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/zalgonoise/x/audio/fft"
	"github.com/zalgonoise/x/audio/sdk/audio"
	"github.com/zalgonoise/x/audio/sdk/audio/compactors"
	"github.com/zalgonoise/x/audio/sdk/audio/consumers/httpaudio"
	"github.com/zalgonoise/x/audio/sdk/audio/exporters"
	"github.com/zalgonoise/x/audio/sdk/audio/exporters/prom"
	"github.com/zalgonoise/x/audio/sdk/audio/exporters/stdout"
	"github.com/zalgonoise/x/audio/sdk/audio/processors"
	"github.com/zalgonoise/x/audio/sdk/audio/registries/batchreg"
	"github.com/zalgonoise/x/cfg"
)

const (
	shutdownTimeout = 15 * time.Second
)

func main() {
	err, code := run()
	if err != nil {
		slog.Error(
			"audio: runtime error",
			slog.String("error", err.Error()),
		)
	}

	os.Exit(code)
}

func run() (error, int) {
	logHandler := slog.NewTextHandler(os.Stderr, nil)
	logger := slog.New(logHandler)
	ctx := context.Background()

	config, err := NewConfig()
	if err != nil {
		return err, 1
	}

	logger.InfoContext(ctx, "setting up consumer")

	consumer, err := httpaudio.New(
		httpaudio.WithTarget(config.Input),
		httpaudio.WithTimeout(config.Duration),
	)
	if err != nil {
		return err, 1
	}

	logger.InfoContext(ctx, "setting up exporter")
	exporter, err := newExporter(ctx, config, logHandler)
	if err != nil {
		return err, 1
	}

	logger.InfoContext(ctx, "setting up processor")
	proc := processors.PCM(exporter)

	ctx, cancel := context.WithTimeout(context.Background(), config.Duration)
	defer cancel()

	logger.InfoContext(ctx, "reading from consumer")
	reader, err := consumer.Consume(ctx)
	if err != nil {
		return err, 1
	}

	logger.InfoContext(ctx, "processing signal")
	go proc.Process(ctx, reader)

	errs := proc.Err()

	defer audio.Shutdown(ctx, shutdownTimeout, consumer, proc)

	for {
		select {
		case <-ctx.Done():
			logger.InfoContext(ctx, "exiting")

			return nil, 0
		case err, ok := <-errs:
			if !ok || err == nil || errors.Is(err, audio.ErrHaltSignal) {
				return nil, 0
			}

			return err, 1
		}
	}
}

func newExporter(ctx context.Context, config *Config, logHandler slog.Handler) (audio.Exporter, error) {
	var (
		logger       = slog.New(logHandler)
		exporterOpts = newExporterOpts(config, logHandler)
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

		exporter, err = prom.ToProm(port, exporterOpts...)
		if err != nil {
			return audio.NoOpExporter(), err
		}
	default:
		exporter, err = stdout.ToLogger(exporterOpts...)
		if err != nil {
			return audio.NoOpExporter(), err
		}
	}

	return exporter, nil
}

func newExporterOpts(config *Config, logHandler slog.Handler) []cfg.Option[exporters.Config] {
	return []cfg.Option[exporters.Config]{
		exporters.WithLogHandler(logHandler),
		newPeaksOpt(config),
		newSpectrumOpt(config),
	}
}

func newPeaksOpt(config *Config) cfg.Option[exporters.Config] {
	if config.Mode == "spectrum" {
		// no-op
		return cfg.Register[exporters.Config](func(config exporters.Config) exporters.Config {
			return config
		})
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

func newSpectrumOpt(config *Config) cfg.Option[exporters.Config] {
	if config.Mode == "peaks" {
		// no-op
		return cfg.Register[exporters.Config](func(config exporters.Config) exporters.Config {
			return config
		})
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
	split := strings.Split(addr, ":")

	switch len(split) {
	case 1:
		port, err = strconv.Atoi(split[0])
		if err != nil {
			return defaultPort, err
		}

		return port, nil
	case 2:
		port, err = strconv.Atoi(split[1])
		if err != nil {
			return defaultPort, err
		}

		return port, nil
	default:
		return defaultPort, nil
	}
}
