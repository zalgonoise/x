package exporters

import (
	"time"

	"github.com/zalgonoise/cfg"
)

type SQLiteConfig struct {
	dur time.Duration
}

func defaultConfig() SQLiteConfig {
	return SQLiteConfig{dur: defaultDur}
}

const (
	minDur     = 5 * time.Second
	defaultDur = time.Minute
)

func WithFlushDuration(dur time.Duration) cfg.Option[SQLiteConfig] {
	if dur < minDur {
		return cfg.NoOp[SQLiteConfig]{}
	}

	return cfg.Register[SQLiteConfig](func(config SQLiteConfig) SQLiteConfig {
		config.dur = dur

		return config
	})
}
