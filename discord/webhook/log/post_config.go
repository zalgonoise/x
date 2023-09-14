package httplog

import (
	"log/slog"
	"time"

	"github.com/zalgonoise/x/cfg"
)

type Config struct {
	level   slog.Leveler
	timeout time.Duration
}

func WithLevel(level slog.Leveler) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.level = level

		return config
	})
}

func WithTimeout(dur time.Duration) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.timeout = dur

		return config
	})
}
