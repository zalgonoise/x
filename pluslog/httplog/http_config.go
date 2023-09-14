package httplog

import (
	"log/slog"
	"time"

	"github.com/zalgonoise/x/cfg"
)

type Config struct {
	source  bool
	encoder Encoder
	level   slog.Leveler
	timeout time.Duration
}

func WithSource() cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.source = true

		return config
	})
}

func WithEncoder(enc Encoder) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.encoder = enc

		return config
	})
}

func JSON(indent bool) cfg.Option[Config] {
	return cfg.Register(func(config Config) Config {
		config.encoder = MarshalJSON{
			indent: indent,
		}

		return config
	})
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
