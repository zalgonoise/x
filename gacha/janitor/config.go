package janitor

import (
	"time"

	"github.com/zalgonoise/cfg"

	"github.com/zalgonoise/x/gacha/poller"
)

const (
	defaultInterval  = time.Hour
	defaultBatchSize = 1000
	defaultLease     = 5 * time.Minute
)

// Config holds the configuration of a janitor.
type Config struct {
	interval  time.Duration
	batchSize int
	lease     time.Duration
	extra     []cfg.Option[poller.Config]
}

func defaultConfig() Config {
	return Config{
		interval:  defaultInterval,
		batchSize: defaultBatchSize,
		lease:     defaultLease,
	}
}

// WithInterval sets how often the janitor runs, and how long an entry whose archive failed waits before the next
// attempt. Each run drains the view, in full batches, before waiting again. It defaults to one hour.
func WithInterval(d time.Duration) cfg.Option[Config] {
	if d <= 0 {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(c Config) Config {
		c.interval = d

		return c
	})
}

// WithBatchSize sets the maximum number of entries archived and deleted per batch. It defaults to 1000.
func WithBatchSize(size int) cfg.Option[Config] {
	if size <= 0 {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(c Config) Config {
		c.batchSize = size

		return c
	})
}

// WithLease sets how long a batch stays leased to the janitor while it is archived. It defaults to five minutes.
func WithLease(d time.Duration) cfg.Option[Config] {
	if d <= 0 {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(c Config) Config {
		c.lease = d

		return c
	})
}

// WithPollerOptions passes options to the underlying poller, such as poller.WithLogger, poller.WithObserver or
// poller.WithConcurrency. They are applied after the janitor's own settings, so they can override them.
func WithPollerOptions(opts ...cfg.Option[poller.Config]) cfg.Option[Config] {
	return cfg.Register(func(c Config) Config {
		c.extra = append(c.extra, opts...)

		return c
	})
}
