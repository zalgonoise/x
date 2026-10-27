package memory

import "github.com/zalgonoise/cfg"

// Config holds the configuration of a memory Repository.
type Config struct {
	deleteOnComplete bool
}

func defaultConfig() Config {
	return Config{}
}

// WithDeleteOnComplete makes Settle delete completed entries instead of keeping them for the janitor. Use it for
// streams that need no audit trail. The Completed view is then always empty.
func WithDeleteOnComplete() cfg.Option[Config] {
	return cfg.Register(func(c Config) Config {
		c.deleteOnComplete = true

		return c
	})
}
