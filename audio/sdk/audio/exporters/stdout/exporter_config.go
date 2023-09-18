package stdout

import (
	"log/slog"

	"github.com/zalgonoise/x/audio/errs"
	"github.com/zalgonoise/x/audio/validation"
	"github.com/zalgonoise/x/cfg"
)

const (
	consumerDomain = errs.Domain("audio/sdk/audio/exporters/stdout")

	ErrTiny = errs.Kind("tiny")

	ErrBlockSize = errs.Entity("spectrum block size")
)

var (
	ErrTinyBlockSize = errs.New(consumerDomain, ErrTiny, ErrBlockSize)

	configValidator = validation.Register[Config](validateSpectrumBlockSize)
	defaultConfig   = Config{}
)

type Config struct {
	withPeaks         bool
	withSpectrum      bool
	spectrumBlockSize int

	logger  *slog.Logger
	handler slog.Handler
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
			blockSize = 64
		}

		config.spectrumBlockSize = blockSize

		return config
	})
}

func validateSpectrumBlockSize(config Config) error {
	if config.withSpectrum {
		if config.spectrumBlockSize < 8 {
			return ErrTinyBlockSize
		}
	}

	return nil
}

func Validate(config Config) error {
	return configValidator.Validate(config)
}

func WithLogger(logger *slog.Logger) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.logger = logger

		return config
	})
}

func WithHandler(h slog.Handler) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.handler = h

		return config
	})
}
