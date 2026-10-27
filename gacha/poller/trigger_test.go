package poller_test

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/gacha/poller"
)

// waitFor runs trigger.Wait and returns how long it took.
func waitFor(ctx context.Context, t *testing.T, trigger poller.Trigger, last poller.Result) (time.Duration, error) {
	t.Helper()

	start := time.Now()
	err := trigger.Wait(ctx, last)

	return time.Since(start), err
}

func TestTicker(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		trigger := poller.Ticker(time.Second)

		elapsed, err := waitFor(t.Context(), t, trigger, poller.Result{})
		require.NoError(t, err)
		require.Equal(t, time.Second, elapsed)

		// A poll that outlasts the interval is followed by another one right away.
		time.Sleep(1500 * time.Millisecond)

		elapsed, err = waitFor(t.Context(), t, trigger, poller.Result{})
		require.NoError(t, err)
		require.Zero(t, elapsed)

		elapsed, err = waitFor(t.Context(), t, trigger, poller.Result{})
		require.NoError(t, err)
		require.Equal(t, 500*time.Millisecond, elapsed, "the ticker keeps a fixed rate")

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		_, err = waitFor(ctx, t, trigger, poller.Result{})
		require.ErrorIs(t, err, context.Canceled)
	})
}

func TestBackoff(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		trigger := poller.Backoff(100*time.Millisecond, time.Second)
		idle := poller.Result{}
		busy := poller.Result{Claimed: 1}

		for i, step := range []struct {
			last  poller.Result
			wants time.Duration
		}{
			{last: idle, wants: 100 * time.Millisecond},
			{last: idle, wants: 200 * time.Millisecond},
			{last: idle, wants: 400 * time.Millisecond},
			{last: idle, wants: 800 * time.Millisecond},
			{last: idle, wants: time.Second},
			{last: idle, wants: time.Second},
			{last: busy, wants: 100 * time.Millisecond},
			{last: idle, wants: 200 * time.Millisecond},
		} {
			elapsed, err := waitFor(t.Context(), t, trigger, step.last)
			require.NoError(t, err)
			require.Equal(t, step.wants, elapsed, "step %d", i)
		}

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		_, err := waitFor(ctx, t, trigger, idle)
		require.ErrorIs(t, err, context.Canceled)
	})
}

func TestNotifier(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		notifier := poller.NewNotifier()

		// Notifications sent before Wait are kept, and coalesced.
		notifier.Notify()
		notifier.Notify()

		elapsed, err := waitFor(t.Context(), t, notifier, poller.Result{})
		require.NoError(t, err)
		require.Zero(t, elapsed)

		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()

		elapsed, err = waitFor(ctx, t, notifier, poller.Result{})
		require.ErrorIs(t, err, context.DeadlineExceeded)
		require.Equal(t, time.Second, elapsed)

		go func() {
			time.Sleep(time.Minute)
			notifier.Notify()
		}()

		elapsed, err = waitFor(t.Context(), t, notifier, poller.Result{})
		require.NoError(t, err)
		require.Equal(t, time.Minute, elapsed)
	})
}

func TestAny(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		notifier := poller.NewNotifier()
		trigger := poller.Any(notifier, poller.Ticker(time.Hour))

		notifier.Notify()

		elapsed, err := waitFor(t.Context(), t, trigger, poller.Result{})
		require.NoError(t, err)
		require.Zero(t, elapsed, "the notifier fires first")

		elapsed, err = waitFor(t.Context(), t, trigger, poller.Result{})
		require.NoError(t, err)
		require.Equal(t, time.Hour, elapsed, "the ticker is the safety net")

		ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
		defer cancel()

		elapsed, err = waitFor(ctx, t, poller.Any(), poller.Result{})
		require.ErrorIs(t, err, context.DeadlineExceeded)
		require.Equal(t, time.Minute, elapsed)

		failing := poller.TriggerFunc(func(context.Context, poller.Result) error {
			return errBoom
		})

		_, err = waitFor(t.Context(), t, poller.Any(notifier, failing), poller.Result{})
		require.ErrorIs(t, err, errBoom)
	})
}

func TestBackoff_ZeroMinimum(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		trigger := poller.Backoff(0, 4*time.Millisecond)
		idle := poller.Result{}

		for i, wants := range []time.Duration{0, time.Millisecond, 2 * time.Millisecond, 4 * time.Millisecond} {
			elapsed, err := waitFor(t.Context(), t, trigger, idle)
			require.NoError(t, err)
			require.Equal(t, wants, elapsed, "step %d", i)
		}

		elapsed, err := waitFor(t.Context(), t, trigger, poller.Result{Claimed: 1})
		require.NoError(t, err)
		require.Zero(t, elapsed, "work found: no wait")
	})
}

func TestAny_IgnoresNil(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		notifier := poller.NewNotifier()
		notifier.Notify()

		elapsed, err := waitFor(t.Context(), t, poller.Any(nil, notifier, nil), poller.Result{})
		require.NoError(t, err)
		require.Zero(t, elapsed)
	})
}
