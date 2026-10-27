// Package writer provides the gacha.Writer implementations: Sync, which writes each call's items directly, and
// Batch, which coalesces concurrent calls into shared writes (group commit).
//
// Both marshal items with a gacha.Marshaler and write the raw values to a gacha.Sink, which can be a repository, a
// sink bound to the caller's transaction, or any gacha.SinkFunc. Both return only once the items are persisted.
package writer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/zalgonoise/x/gacha"
)

// marshal appends the raw form of items to raws.
func marshal[T, R any](m gacha.Marshaler[T, R], raws []R, items []T) ([]R, error) {
	for i := range items {
		raw, err := m.Marshal(items[i])
		if err != nil {
			return raws, fmt.Errorf("%w %d: %w", ErrMarshal, i, err)
		}

		raws = append(raws, raw)
	}

	return raws, nil
}

// write writes data to sink, retrying failures as configured.
func write[R any](ctx context.Context, sink gacha.Sink[R], data []R, config *Config) error {
	for retry := 1; ; retry++ {
		err := sink.Write(ctx, data...)
		if err == nil {
			return nil
		}

		if retry > config.retries || !retryable(err) {
			return err
		}

		if config.backoff != nil {
			if waitErr := sleep(ctx, config.backoff(retry)); waitErr != nil {
				return errors.Join(err, waitErr)
			}
		}
	}
}

func retryable(err error) bool {
	return !gacha.IsPermanent(err) && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded)
}

func sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}

	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func notify(fns []func()) {
	for _, fn := range fns {
		fn()
	}
}
