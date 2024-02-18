package tracing

import (
	"time"

	"github.com/zalgonoise/cfg"
)

const defaultTimeout = 2 * time.Minute

type Config struct {
	username string
	password string
	timeout  time.Duration
}

func defaultConfig() Config {
	return Config{
		timeout: defaultTimeout,
	}
}

func WithCredentials(username, password string) cfg.Option[Config] {
	if username == "" && password == "" {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register[Config](func(config Config) Config {
		config.username = username
		config.password = password

		return config
	})
}

func WithTimeout(dur time.Duration) cfg.Option[Config] {
	if dur <= 0 {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register[Config](func(config Config) Config {
		config.timeout = dur

		return config
	})
}
