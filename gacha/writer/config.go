package writer

import (
	"time"

	"github.com/zalgonoise/cfg"
)

const defaultBatchMaxSize = 1000

// Config holds the configuration of the Sync and Batch writers.
type Config struct {
	maxBatchSize int
	linger       time.Duration
	flushTimeout time.Duration
	retries      int
	backoff      func(retry int) time.Duration
	notify       []func()
}

func defaultSyncConfig() Config {
	return Config{}
}

func defaultBatchConfig() Config {
	return Config{
		maxBatchSize: defaultBatchMaxSize,
	}
}

// WithMaxBatchSize caps the number of raw items per Sink.Write call. Zero or a negative size removes the cap.
//
// Sync splits larger inputs into chunks of size items, written one after the other: a failure leaves the earlier
// chunks persisted. Sync has no cap by default.
//
// Batch flushes as soon as the coalesced writes reach size items. A single Write larger than size is flushed on its
// own, never split. Batch defaults to 1000 items.
func WithMaxBatchSize(size int) cfg.Option[Config] {
	return cfg.Register(func(c Config) Config {
		c.maxBatchSize = max(size, 0)

		return c
	})
}

// WithLinger makes Batch wait up to d after the first write of a batch for more writes to coalesce with, unless the
// batch fills up first. By default, Batch does not wait: it flushes whatever is queued as soon as the previous flush
// returns, so writes coalesce only while a flush is in flight. Sync ignores this option.
func WithLinger(d time.Duration) cfg.Option[Config] {
	return cfg.Register(func(c Config) Config {
		c.linger = max(d, 0)

		return c
	})
}

// WithFlushTimeout bounds each Batch flush, retries included, to d. Flushes run detached from the callers'
// contexts, since they serve several callers at once. There is no timeout by default. Sync ignores this option: it
// uses the caller's context.
func WithFlushTimeout(d time.Duration) cfg.Option[Config] {
	return cfg.Register(func(c Config) Config {
		c.flushTimeout = max(d, 0)

		return c
	})
}

// WithRetry retries a failed Sink.Write up to retries more times, waiting backoff(n) before the n-th retry (n starts
// at 1). A nil backoff retries right away. Errors marked with gacha.Permanent and context errors are not retried.
// Functions from the poller package, such as poller.Exponential, can serve as backoff.
//
// A retried write may persist the items twice when a failed attempt had in fact committed (for instance, when the
// connection dropped after the commit). The duplicates get distinct entry IDs, so handlers that must be idempotent
// across write retries should key on the item's own identity.
//
// Do not retry sinks bound to a transaction (such as a repository's Tx sink): a failed statement may abort the
// transaction (PostgreSQL does), and only retrying the whole transaction helps then.
func WithRetry(retries int, backoff func(retry int) time.Duration) cfg.Option[Config] {
	return cfg.Register(func(c Config) Config {
		c.retries = max(retries, 0)
		c.backoff = backoff

		return c
	})
}

// WithNotify registers functions called after each successful write, such as poller.Notifier.Notify, to wake a
// poller as soon as there is work. They are called synchronously, so they should return quickly.
func WithNotify(fns ...func()) cfg.Option[Config] {
	return cfg.Register(func(c Config) Config {
		for _, fn := range fns {
			if fn != nil {
				c.notify = append(c.notify, fn)
			}
		}

		return c
	})
}
