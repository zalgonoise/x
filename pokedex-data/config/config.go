package config

import "github.com/zalgonoise/cfg"

const (
	minID = 1
	maxID = 1025
)

type Config struct {
	Output string
	Min    int
	Max    int
}

func DefaultConfig() Config {
	return Config{
		Output: "pokemon.csv",
		Min:    minID,
		Max:    maxID,
	}
}

func WithOutput(path string) cfg.Option[Config] {
	if path == "" {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register[Config](func(config Config) Config {
		config.Output = path

		return config
	})
}

func WithMin(min int) cfg.Option[Config] {
	if min < minID || min > maxID {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register[Config](func(config Config) Config {
		config.Min = min

		return config
	})
}

func WithMax(max int) cfg.Option[Config] {
	if max < minID || max > maxID {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register[Config](func(config Config) Config {
		config.Max = max

		return config
	})
}

func Merge(base, next Config) Config {
	if next.Output != "" {
		base.Output = next.Output
	}

	if next.Min > 0 {
		base.Min = next.Min
	}

	if next.Max > 0 {
		base.Max = next.Max
	}

	return base
}
