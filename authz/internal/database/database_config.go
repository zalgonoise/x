package database

import (
	"github.com/zalgonoise/cfg"
)

type Config struct {
	maxOpenConns int
	maxIdleConns int
}

func WithMaxConns(open, idle int) cfg.Option[Config] {
	if open <= 0 {
		open = defaultMaxOpenConns
	}

	if idle <= 0 {
		idle = defaultMaxIdleConns
	}

	return cfg.Register[Config](func(config Config) Config {
		config.maxOpenConns = open
		config.maxIdleConns = idle

		return config
	})
}
