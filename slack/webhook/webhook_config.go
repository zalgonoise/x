package webhook

import (
	"log/slog"
	"time"

	"github.com/zalgonoise/cfg"
)

type Config struct {
	timeout time.Duration

	handler slog.Handler
}

func WithTimeout(timeout time.Duration) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.timeout = timeout

		return config
	})
}

func WithLogger(logger *slog.Logger) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.handler = logger.Handler()

		return config
	})
}

func WithLogHandler(handler slog.Handler) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.handler = handler

		return config
	})
}
