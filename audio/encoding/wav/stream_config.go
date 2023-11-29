package wav

import (
	"time"

	"github.com/zalgonoise/cfg"
)

// Config holds different configuration settings for a Stream.
type Config struct {
	size  int
	dur   time.Duration
	ratio float64
}

// WithSize defines a concrete value for the Stream's buffer size (in bytes).
func WithSize(size int) cfg.Option[Config] {
	if size <= 0 {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(config Config) Config {
		config.size = size

		return config
	})
}

// WithDuration sets a time.Duration value for the desired Stream buffer, that is later translated to a concrete
// (byte-size) value.
func WithDuration(dur time.Duration) cfg.Option[Config] {
	if dur <= 0 {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(config Config) Config {
		config.dur = dur

		return config
	})
}

// WithRatio sets a float64 value describing a ratio against 1 second (e.g. 0.5 is half-a-second, 2.0 is two seconds).
func WithRatio(ratio float64) cfg.Option[Config] {
	if ratio <= 0.0 {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(config Config) Config {
		config.ratio = ratio

		return config
	})
}
