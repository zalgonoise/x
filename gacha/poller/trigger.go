package poller

import (
	"context"
	"sync"
	"time"
)

// Trigger decides when the poller claims the next batch.
//
// A Trigger instance serves one poller: pollers sharing a Ticker, a Notifier or a Listener split its wake-ups between
// them. Give each poller its own, or combine per-poller Notifiers with writer.WithNotify.
type Trigger interface {
	// Wait blocks until the next poll should start, or until ctx is done (returning the context's error). last is the
	// result of the previous poll. Any other error stops the poller's Run, which returns it.
	Wait(ctx context.Context, last Result) error
}

// TriggerFunc is a function type that implements Trigger.
type TriggerFunc func(ctx context.Context, last Result) error

// Wait calls fn(ctx, last).
func (fn TriggerFunc) Wait(ctx context.Context, last Result) error {
	return fn(ctx, last)
}

type ticker struct {
	mu     sync.Mutex
	d      time.Duration
	ticker *time.Ticker
}

// Ticker returns a Trigger that fires at a fixed interval d. When a poll takes longer than d, the next one starts
// right away. The interval starts with the first Wait.
func Ticker(d time.Duration) Trigger {
	return &ticker{d: max(d, time.Nanosecond)}
}

func (t *ticker) Wait(ctx context.Context, _ Result) error {
	t.mu.Lock()
	if t.ticker == nil {
		t.ticker = time.NewTicker(t.d)
	}

	c := t.ticker.C
	t.mu.Unlock()

	select {
	case <-c:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

const (
	backoffFactor = 2
	// backoffFloor is the shortest idle wait of a Backoff, so that Backoff(0, max) still backs off.
	backoffFloor = time.Millisecond
)

type backoff struct {
	mu       sync.Mutex
	min, max time.Duration
	current  time.Duration
	started  bool
}

// Backoff returns an adaptive Trigger: it waits minWait after a poll that claimed entries, and doubles the wait after
// every poll that claimed none (or failed), up to maxWait. Idle waits are at least a millisecond, even when minWait
// is zero.
func Backoff(minWait, maxWait time.Duration) Trigger {
	minWait = max(minWait, 0)

	return &backoff{min: minWait, max: max(maxWait, minWait, backoffFloor)}
}

func (b *backoff) Wait(ctx context.Context, last Result) error {
	b.mu.Lock()

	switch {
	case last.Claimed > 0 || !b.started:
		b.current = b.min
		b.started = true
	default:
		b.current = min(max(b.current*backoffFactor, backoffFloor), b.max)
	}

	d := b.current
	b.mu.Unlock()

	return sleep(ctx, d)
}

// Notifier is an event-driven Trigger: Wait returns when Notify is called. Notifications sent while no poll is
// waiting are coalesced into one, so a single Notify after a batch of writes is enough.
//
// Pass Notify to writer.WithNotify to poll as soon as entries are written, and combine the Notifier with a Ticker
// through Any as a safety net for entries written elsewhere and for retries.
type Notifier struct {
	ch chan struct{}
}

// NewNotifier creates a Notifier.
func NewNotifier() *Notifier {
	return &Notifier{ch: make(chan struct{}, 1)}
}

// Notify wakes the poller waiting on the Notifier, or the next one to wait. It never blocks.
func (n *Notifier) Notify() {
	select {
	case n.ch <- struct{}{}:
	default:
	}
}

// Wait blocks until Notify is called or ctx is done.
func (n *Notifier) Wait(ctx context.Context, _ Result) error {
	select {
	case <-n.ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type anyTrigger []Trigger

// Any returns a Trigger that fires as soon as any of triggers fires, such as a Notifier with a Ticker as a safety
// net. Nil triggers are ignored. Without triggers, it waits until the context is done.
func Any(triggers ...Trigger) Trigger {
	nonNil := make(anyTrigger, 0, len(triggers))

	for _, trigger := range triggers {
		if trigger != nil {
			nonNil = append(nonNil, trigger)
		}
	}

	return nonNil
}

func (a anyTrigger) Wait(ctx context.Context, last Result) error {
	if len(a) == 0 {
		<-ctx.Done()

		return ctx.Err()
	}

	var wg sync.WaitGroup

	waitCtx, cancel := context.WithCancel(ctx)
	results := make(chan error, len(a))

	for _, trigger := range a {
		wg.Go(func() {
			results <- trigger.Wait(waitCtx, last)
		})
	}

	err := <-results

	// Stop the other triggers and wait for them, so none outlives this call.
	cancel()
	wg.Wait()

	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}

	return err
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
