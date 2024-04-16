package exporters

import (
	"log/slog"

	"github.com/zalgonoise/cfg"

	"github.com/zalgonoise/x/audio/fft"
	"github.com/zalgonoise/x/audio/sdk/audio/registries/batchreg"
)

const (
	minBlockSize     = 8
	defaultBlockSize = 64
)

func DefaultStatsConfig() *StatsConfig {
	return &StatsConfig{
		spectrumBlockSize: defaultBlockSize,
	}
}

type StatsConfig struct {
	withPeaks         bool
	withSpectrum      bool
	spectrumBlockSize int

	batchedPeaks        bool
	batchedPeaksOptions []cfg.Option[batchreg.Config[float64]]

	batchedSpectrum        bool
	batchedSpectrumOptions []cfg.Option[batchreg.Config[[]fft.FrequencyPower]]

	LogHandler slog.Handler
}

func WithBatchedPeaks(options ...cfg.Option[batchreg.Config[float64]]) cfg.Option[*StatsConfig] {
	return cfg.Register(func(config *StatsConfig) *StatsConfig {
		config.withPeaks = true
		config.batchedPeaks = true
		config.batchedPeaksOptions = options

		return config
	})
}

func WithBatchedSpectrum(
	blockSize int,
	options ...cfg.Option[batchreg.Config[[]fft.FrequencyPower]],
) cfg.Option[*StatsConfig] {
	return cfg.Register(func(config *StatsConfig) *StatsConfig {
		if blockSize < minBlockSize {
			blockSize = defaultBlockSize
		}

		config.withSpectrum = true
		config.spectrumBlockSize = blockSize
		config.batchedSpectrum = true
		config.batchedSpectrumOptions = options

		return config
	})
}

func WithPeaks() cfg.Option[*StatsConfig] {
	return cfg.Register(func(config *StatsConfig) *StatsConfig {
		config.withPeaks = true

		return config
	})
}

func WithSpectrum(blockSize int) cfg.Option[*StatsConfig] {
	return cfg.Register(func(config *StatsConfig) *StatsConfig {
		config.withSpectrum = true

		if blockSize < minBlockSize {
			blockSize = defaultBlockSize
		}

		config.spectrumBlockSize = blockSize

		return config
	})
}

func WithLogger(logger *slog.Logger) cfg.Option[*StatsConfig] {
	if logger == nil {
		return cfg.NoOp[*StatsConfig]{}
	}

	return cfg.Register(func(config *StatsConfig) *StatsConfig {
		config.LogHandler = logger.Handler()

		return config
	})
}

func WithLogHandler(h slog.Handler) cfg.Option[*StatsConfig] {
	if h == nil {
		return cfg.NoOp[*StatsConfig]{}
	}

	return cfg.Register(func(config *StatsConfig) *StatsConfig {
		config.LogHandler = h

		return config
	})
}
