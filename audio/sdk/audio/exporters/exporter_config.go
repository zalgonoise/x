package exporters

import (
	"log/slog"

	"github.com/zalgonoise/x/audio/fft"
	"github.com/zalgonoise/x/audio/sdk/audio/registries/batchreg"
	"github.com/zalgonoise/x/cfg"
)

const defaultBlockSize = 64

var DefaultConfig = Config{
	spectrumBlockSize: defaultBlockSize,
}

type Config struct {
	withPeaks         bool
	withSpectrum      bool
	spectrumBlockSize int

	batchedPeaks        bool
	batchedPeaksOptions []cfg.Option[batchreg.Config[float64]]

	batchedSpectrum        bool
	batchedSpectrumOptions []cfg.Option[batchreg.Config[[]fft.FrequencyPower]]

	LogHandler slog.Handler
}

func WithBatchedPeaks(options ...cfg.Option[batchreg.Config[float64]]) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.withPeaks = true
		config.batchedPeaks = true
		config.batchedPeaksOptions = options

		return config
	})
}

func WithBatchedSpectrum(blockSize int, options ...cfg.Option[batchreg.Config[[]fft.FrequencyPower]]) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		if blockSize < 8 {
			blockSize = defaultBlockSize
		}

		config.withSpectrum = true
		config.spectrumBlockSize = blockSize
		config.batchedSpectrum = true
		config.batchedSpectrumOptions = options

		return config
	})
}

func WithPeaks() cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.withPeaks = true

		return config
	})
}

func WithSpectrum(blockSize int) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.withSpectrum = true

		if blockSize < 8 {
			blockSize = defaultBlockSize
		}

		config.spectrumBlockSize = blockSize

		return config
	})
}

func WithLogger(logger *slog.Logger) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.LogHandler = logger.Handler()

		return config
	})
}

func WithLogHandler(h slog.Handler) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.LogHandler = h

		return config
	})
}