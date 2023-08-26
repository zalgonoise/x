package httpaudio

import (
	"time"

	"github.com/zalgonoise/x/cfg"
)

// HTTPConfig defines a data structure for configurations and options related to a HTTP audio.Consumer
type HTTPConfig struct {
	target  string
	timeout time.Duration
}

// WithTimeout sets a general timeout for the HTTP connection.
func WithTimeout(dur time.Duration) cfg.Option[HTTPConfig] {
	return cfg.Register(func(config HTTPConfig) HTTPConfig {
		config.timeout = dur

		return config
	})
}

// WithTarget defines the HTTP URL of the audio source.
func WithTarget(target string) cfg.Option[HTTPConfig] {
	return cfg.Register(func(config HTTPConfig) HTTPConfig {
		config.target = target

		return config
	})
}

func newHTTPConfig(options ...cfg.Option[HTTPConfig]) HTTPConfig {
	return cfg.New[HTTPConfig](options...)
}
