package stdout

import (
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

	configValidator = validation.Register[LoggerConfig](validateSpectrumBlockSize)
	defaultConfig   = LoggerConfig{}
)

type LoggerConfig struct {
	withPeaks bool

	withSpectrum      bool
	spectrumBlockSize int
}

func WithPeaks() cfg.Option[LoggerConfig] {
	return cfg.Register(func(config LoggerConfig) LoggerConfig {
		config.withPeaks = true

		return config
	})
}

func WithSpectrum(blockSize int) cfg.Option[LoggerConfig] {
	return cfg.Register(func(config LoggerConfig) LoggerConfig {
		config.withSpectrum = true

		if blockSize < 8 {
			blockSize = 64
		}

		config.spectrumBlockSize = blockSize

		return config
	})
}

func validateSpectrumBlockSize(config LoggerConfig) error {
	if config.spectrumBlockSize < 8 {
		return ErrTinyBlockSize
	}

	return nil
}

func Validate(config LoggerConfig) error {
	return configValidator.Validate(config)
}
