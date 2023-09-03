package schedule

import (
	"time"

	"github.com/zalgonoise/x/cfg"
)

type CronConfig struct {
	cronString string
	loc        *time.Location
}

func WithSchedule(cronString string) cfg.Option[CronConfig] {
	return cfg.Register(func(config CronConfig) CronConfig {
		config.cronString = cronString

		return config
	})
}

func WithLocation(loc *time.Location) cfg.Option[CronConfig] {
	return cfg.Register(func(config CronConfig) CronConfig {
		config.loc = loc

		return config
	})
}
