package exporters

import (
	"time"

	"github.com/zalgonoise/cfg"

	"github.com/zalgonoise/x/audio/sdk/audio"
	"github.com/zalgonoise/x/audio/sdk/audio/extractors"
)

const (
	defaultDuration    = time.Minute
	defaultSampleRate  = 44100
	defaultNumChannels = 2
	defaultBitDepth    = 32

	numSeconds = 60
)

type DataConfig struct {
	sampleRate  uint32
	numChannels uint16
	bitDepth    uint16

	extractor   audio.Extractor[float64]
	maxSamples  int64
	maxDuration time.Duration
	threshold   func(float64) bool
}

func defaultDataConfig() DataConfig {
	return DataConfig{
		sampleRate:  defaultSampleRate,
		numChannels: defaultNumChannels,
		bitDepth:    defaultBitDepth,
		extractor:   extractors.MaxAbsPeak(),
		maxDuration: defaultDuration,
		threshold:   audio.NoOpThreshold[float64](),
	}
}

func AsWAV(sampleRate uint32, numChannels, bitDepth uint16) cfg.Option[DataConfig] {
	return cfg.Register(func(config DataConfig) DataConfig {
		config.sampleRate = sampleRate
		config.numChannels = numChannels
		config.bitDepth = bitDepth

		return config
	})
}

func WithDuration(dur time.Duration) cfg.Option[DataConfig] {
	return cfg.Register(func(config DataConfig) DataConfig {
		config.maxDuration = dur

		return config
	})
}

func WithMaxSamples(samples int64) cfg.Option[DataConfig] {
	return cfg.Register(func(config DataConfig) DataConfig {
		config.maxSamples = samples

		return config
	})
}

func WithExtractor(extractor audio.Extractor[float64]) cfg.Option[DataConfig] {
	return cfg.Register(func(config DataConfig) DataConfig {
		config.extractor = extractor

		return config
	})
}

func WithThreshold(threshold audio.Threshold[float64]) cfg.Option[DataConfig] {
	return cfg.Register(func(config DataConfig) DataConfig {
		config.threshold = threshold

		return config
	})
}
