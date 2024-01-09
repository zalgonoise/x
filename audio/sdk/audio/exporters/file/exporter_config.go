package file

import "github.com/zalgonoise/cfg"

type Config struct {
	sampleRate     uint32
	numChannels    uint16
	bitDepth       uint16
	outputDir      string
	filenamePrefix string
}

func defaultConfig() Config {
	return Config{
		sampleRate:  44100,
		numChannels: 2,
		bitDepth:    32,
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

func WithOutputDir(path string) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.outputDir = path

		return config
	})
}

func WithNamePrefix(name string) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.filenamePrefix = name

		return config
	})
}
