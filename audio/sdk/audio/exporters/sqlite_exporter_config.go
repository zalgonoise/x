package exporters

import "github.com/zalgonoise/cfg"

type SQLiteConfig struct {
	size int
}

func defaultConfig() SQLiteConfig {
	return SQLiteConfig{size: defaultSize}
}

const minSize = 1 << 12

func WithFlushSize(size int) cfg.Option[SQLiteConfig] {
	if size < minSize {
		return cfg.NoOp[SQLiteConfig]{}
	}

	return cfg.Register[SQLiteConfig](func(config SQLiteConfig) SQLiteConfig {
		config.size = size

		return config
	})
}
