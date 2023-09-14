package webhook

import (
	"github.com/zalgonoise/x/cfg"
	"log/slog"
	"time"
)

type Config struct {
	timeout time.Duration

	logger *slog.Logger
}

func WithTimeout(timeout time.Duration) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.timeout = timeout

		return config
	})
}

func WithLogger(logger *slog.Logger) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.logger = logger

		return config
	})
}
