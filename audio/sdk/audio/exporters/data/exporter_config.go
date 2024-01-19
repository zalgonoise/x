package data

import (
	"time"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/audio/sdk/audio"
	"github.com/zalgonoise/x/audio/sdk/audio/extractors"
)

const (
	defaultDuration  = time.Minute
	numSeconds       = 60
	defaultThreshold = 0.4
)

type Config struct {
	sampleRate  uint32
	numChannels uint16
	bitDepth    uint16

	extractor   audio.Extractor[float64]
	maxSamples  int64
	maxDuration time.Duration
	threshold   func(float64) bool
}

func defaultConfig() Config {
	return Config{
		sampleRate:  44100,
		numChannels: 2,
		bitDepth:    32,
		extractor:   extractors.MaxPeak(),
		maxDuration: defaultDuration,
		threshold: func(f float64) bool {
			return f >= defaultThreshold
		},
	}
}

func AsWAV(sampleRate uint32, numChannels, bitDepth uint16) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.sampleRate = sampleRate
		config.numChannels = numChannels
		config.bitDepth = bitDepth

		return config
	})
}

func WithDuration(dur time.Duration) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.maxDuration = dur

		return config
	})
}

func WithMaxSamples(samples int64) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.maxSamples = samples

		return config
	})
}

func WithExtractor(extractor audio.Extractor[float64]) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.extractor = extractor

		return config
	})
}

func WithThreshold(threshold func(float64) bool) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.threshold = threshold

		return config
	})
}
