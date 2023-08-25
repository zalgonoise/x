package httpaudio

import (
	"time"

	"github.com/zalgonoise/x/audio/sdk/audio"
)

// HTTPConfig defines a data structure for configurations and options related to a HTTP audio.Consumer
type HTTPConfig struct {
	target  string
	timeout time.Duration
}

// WithTimeout sets a general timeout for the HTTP connection.
func WithTimeout(dur time.Duration) audio.Option[HTTPConfig] {
	return audio.Register(func(config HTTPConfig) HTTPConfig {
		config.timeout = dur

		return config
	})
}

// WithTarget defines the HTTP URL of the audio source.
func WithTarget(target string) audio.Option[HTTPConfig] {
	return audio.Register(func(config HTTPConfig) HTTPConfig {
		config.target = target

		return config
	})
}

func newHTTPConfig(options ...audio.Option[HTTPConfig]) HTTPConfig {
	return audio.NewConfig[HTTPConfig](options...)
}
