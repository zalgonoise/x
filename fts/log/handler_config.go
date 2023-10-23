package log

import (
	"log/slog"

	"github.com/zalgonoise/x/cfg"
)

type Config struct {
	withSpanID bool
	handler    slog.Handler
}

func WithSpanID() cfg.Option[Config] {
	return cfg.Register[Config](func(config Config) Config {
		config.withSpanID = true

		return config
	})
}

func WithHandler(handler slog.Handler) cfg.Option[Config] {
	if handler == nil {
		handler = defaultHandler()
	}

	return cfg.Register[Config](func(config Config) Config {
		config.handler = handler

		return config
	})
}

func WithLogger(logger *slog.Logger) cfg.Option[Config] {
	if logger == nil {
		return WithHandler(defaultHandler())
	}

	return cfg.Register[Config](func(config Config) Config {
		config.handler = logger.Handler()

		return config
	})
}
