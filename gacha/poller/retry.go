package poller

import (
	"math/rand/v2"
	"time"
)

// Constant returns a retry backoff, for WithRetryBackoff or writer.WithRetry, that always waits d.
func Constant(d time.Duration) func(retry int) time.Duration {
	d = max(d, 0)

	return func(int) time.Duration {
		return d
	}
}

// Exponential returns a retry backoff, for WithRetryBackoff or writer.WithRetry, that waits base before the first
// retry and doubles the delay on every retry after that, up to limit.
func Exponential(base, limit time.Duration) func(retry int) time.Duration {
	base = max(base, 0)
	limit = max(limit, base)

	return func(retry int) time.Duration {
		d := base

		for i := 1; i < retry && d < limit; i++ {
			if d > limit/2 {
				return limit
			}

			d *= 2
		}

		return min(d, limit)
	}
}

// Jitter returns a retry backoff that randomly shortens the delays of backoff by up to factor (between 0 and 1) of
// their length, so that entries failing together do not retry together. For instance, a factor of 0.2 turns a 10s
// delay into a random delay between 8s and 10s.
func Jitter(backoff func(retry int) time.Duration, factor float64) func(retry int) time.Duration {
	factor = min(max(factor, 0), 1)

	return func(retry int) time.Duration {
		d := backoff(retry)

		//nolint:gosec // Jitter does not need a cryptographically secure source.
		return d - time.Duration(float64(d)*factor*rand.Float64())
	}
}
