package poller

import (
	"log/slog"
	"time"

	"github.com/zalgonoise/cfg"
)

const (
	defaultBatchSize     = 100
	defaultLease         = time.Minute
	defaultMaxAttempts   = 10
	defaultSourceTimeout = 10 * time.Second
	defaultWaitFloor     = 100 * time.Millisecond
	defaultWaitCeiling   = 5 * time.Second
	defaultRetryBase     = time.Second
	defaultRetryLimit    = 10 * time.Minute
	defaultRetryJitter   = 0.2
	heartbeatsPerLease   = 3
)

// Config holds the configuration of a Poller.
type Config struct {
	trigger        Trigger
	batchSize      int
	concurrency    int
	lease          time.Duration
	heartbeat      time.Duration
	heartbeatSet   bool
	maxAttempts    int
	retryBackoff   func(retry int) time.Duration
	handlerTimeout time.Duration
	sourceTimeout  time.Duration
	batchSettle    bool
	drain          bool
	observer       func(Result)
	logger         *slog.Logger
}

func defaultConfig() Config {
	return Config{
		batchSize:     defaultBatchSize,
		concurrency:   1,
		lease:         defaultLease,
		maxAttempts:   defaultMaxAttempts,
		retryBackoff:  Jitter(Exponential(defaultRetryBase, defaultRetryLimit), defaultRetryJitter),
		sourceTimeout: defaultSourceTimeout,
		drain:         true,
	}
}

// finish fills in the defaults that depend on other settings, or that must not be shared between pollers.
func (c Config) finish() Config {
	if c.trigger == nil {
		c.trigger = Backoff(defaultWaitFloor, defaultWaitCeiling)
	}

	// A heartbeat that is not shorter than the lease cannot keep it.
	if !c.heartbeatSet || c.heartbeat >= c.lease {
		c.heartbeat = c.lease / heartbeatsPerLease
	}

	if c.logger == nil {
		c.logger = slog.New(slog.DiscardHandler)
	}

	return c
}

// WithTrigger sets the Trigger deciding when Run polls. It defaults to Backoff(100ms, 5s).
func WithTrigger(trigger Trigger) cfg.Option[Config] {
	if trigger == nil {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(c Config) Config {
		c.trigger = trigger

		return c
	})
}

// WithBatchSize sets the maximum number of entries claimed per poll. It defaults to 100.
func WithBatchSize(size int) cfg.Option[Config] {
	if size <= 0 {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(c Config) Config {
		c.batchSize = size

		return c
	})
}

// WithConcurrency sets how many entries of a batch are handled in parallel. It defaults to 1: entries are handled
// one after the other, in claim order. A poll returns once every entry of its batch is settled, so a slow entry holds
// up the next poll: run several pollers to process batches in parallel, which SKIP LOCKED claims and leases make safe.
func WithConcurrency(n int) cfg.Option[Config] {
	if n <= 0 {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(c Config) Config {
		c.concurrency = n

		return c
	})
}

// WithLease sets how long claimed entries stay leased to this poller. An entry whose lease expires before it is
// settled is claimed again, by this poller or another one. It defaults to one minute.
func WithLease(d time.Duration) cfg.Option[Config] {
	if d <= 0 {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(c Config) Config {
		c.lease = d

		return c
	})
}

// WithHeartbeat sets how often the leases of a batch's unsettled entries are renewed while it is being handled, when
// the Source implements gacha.Extender (a Source wrapping a repository must implement it too). Zero or a negative
// interval disables heartbeats. It defaults to a third of the lease, which also replaces intervals that are not
// shorter than the lease.
func WithHeartbeat(d time.Duration) cfg.Option[Config] {
	return cfg.Register(func(c Config) Config {
		c.heartbeat = max(d, 0)
		c.heartbeatSet = true

		return c
	})
}

// WithMaxAttempts sets the number of attempts after which a failing entry is moved to the dead-letter table. Entries
// claimed more times than that (because handlers crashed the process before settling them) are moved without running
// the handler. Zero or a negative value retries forever, which is what a poller over a repository view needs. It
// defaults to 10.
func WithMaxAttempts(n int) cfg.Option[Config] {
	return cfg.Register(func(c Config) Config {
		c.maxAttempts = max(n, 0)

		return c
	})
}

// WithRetryBackoff sets the delay before an entry that failed its n-th attempt can be claimed again. It defaults to
// Jitter(Exponential(time.Second, 10*time.Minute), 0.2).
func WithRetryBackoff(backoff func(retry int) time.Duration) cfg.Option[Config] {
	if backoff == nil {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(c Config) Config {
		c.retryBackoff = backoff

		return c
	})
}

// WithHandlerTimeout bounds each handler call to d. A handler that times out is retried like a failed one. Handlers
// must return when their context is done: the poller waits for them. There is no timeout by default.
func WithHandlerTimeout(d time.Duration) cfg.Option[Config] {
	return cfg.Register(func(c Config) Config {
		c.handlerTimeout = max(d, 0)

		return c
	})
}

// WithSourceTimeout bounds each call to the Source (Claim, Settle and heartbeat Extend calls) to d. Claims and
// settles run detached from the poll's context, so that a shutdown neither strands a committed claim nor loses the
// outcomes of the interrupted batch. It defaults to 10 seconds.
func WithSourceTimeout(d time.Duration) cfg.Option[Config] {
	if d <= 0 {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(c Config) Config {
		c.sourceTimeout = d

		return c
	})
}

// WithBatchSettle settles a batch's outcomes in one Settle call once every entry is handled, instead of settling each
// entry as soon as its handler returns. It saves round-trips, but a crash mid-batch makes every entry of the batch run
// again, including the ones that had finished.
func WithBatchSettle() cfg.Option[Config] {
	return cfg.Register(func(c Config) Config {
		c.batchSettle = true

		return c
	})
}

// WithDrain sets whether Run polls again right away after a full batch, instead of waiting for its Trigger, which
// drains a backlog as fast as possible. It defaults to true.
func WithDrain(drain bool) cfg.Option[Config] {
	return cfg.Register(func(c Config) Config {
		c.drain = drain

		return c
	})
}

// WithObserver registers a function called with the Result of every poll. It is called synchronously, so it should
// return quickly.
func WithObserver(fn func(Result)) cfg.Option[Config] {
	if fn == nil {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(c Config) Config {
		c.observer = fn

		return c
	})
}

// WithLogger sets the logger for claim, settle and heartbeat failures, recovered panics (at error level) and poll
// results (at debug level). Nothing is logged by default.
func WithLogger(logger *slog.Logger) cfg.Option[Config] {
	if logger == nil {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(c Config) Config {
		c.logger = logger

		return c
	})
}
