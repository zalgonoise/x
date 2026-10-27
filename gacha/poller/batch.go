package poller

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
	"uuid"

	"github.com/zalgonoise/x/gacha"
)

// batch tracks the entries of one poll while they are handled and settled.
type batch[T, R any] struct {
	p        *Poller[T, R]
	entries  []gacha.Entry[R]
	outcomes []gacha.Outcome

	mu      sync.Mutex
	settled []bool
	errs    []error
}

func newBatch[T, R any](p *Poller[T, R], entries []gacha.Entry[R]) *batch[T, R] {
	return &batch[T, R]{
		p:        p,
		entries:  entries,
		outcomes: make([]gacha.Outcome, len(entries)),
		settled:  make([]bool, len(entries)),
	}
}

// handle decides the outcome of every entry, with up to the configured concurrency.
func (b *batch[T, R]) handle(ctx context.Context) {
	workers := min(b.p.config.concurrency, len(b.entries))

	if workers == 1 {
		for i := range b.entries {
			b.handleOne(ctx, i)
		}

		return
	}

	var (
		wg   sync.WaitGroup
		next atomic.Int64
	)

	for range workers {
		wg.Go(func() {
			for {
				i := int(next.Add(1) - 1)
				if i >= len(b.entries) {
					return
				}

				b.handleOne(ctx, i)
			}
		})
	}

	wg.Wait()
}

// handleOne decides the outcome of entry i, and settles it right away unless batch settles are on. Released entries
// are left to the final settle, so that a shutdown hands them all back in one call.
func (b *batch[T, R]) handleOne(ctx context.Context, i int) {
	entry := &b.entries[i]

	if ctx.Err() != nil {
		// Shutting down: hand the entry back without running it.
		b.outcomes[i] = gacha.Outcome{ID: entry.ID, Attempts: entry.Attempts, Status: gacha.StatusReleased}

		return
	}

	b.outcomes[i] = b.p.decide(ctx, entry)

	if b.p.config.batchSettle || b.outcomes[i].Status == gacha.StatusReleased {
		return
	}

	if err := b.p.settle(ctx, b.outcomes[i:i+1]); err != nil {
		// The entry stays unsettled: its lease keeps being renewed, and the final settle retries it.
		b.p.config.logger.WarnContext(ctx, "gacha: settle failed, retrying with the rest of the batch",
			slog.String("id", entry.ID.String()),
			slog.Any("error", err),
		)

		return
	}

	b.mu.Lock()
	b.settled[i] = true
	b.mu.Unlock()
}

// settleRemaining settles every outcome not settled yet in one call: the whole batch with batch settles, otherwise
// the released entries and the entries whose own settle failed.
func (b *batch[T, R]) settleRemaining(ctx context.Context) {
	b.mu.Lock()

	pending := make([]gacha.Outcome, 0, len(b.outcomes))

	for i := range b.outcomes {
		if !b.settled[i] {
			pending = append(pending, b.outcomes[i])
		}
	}

	b.mu.Unlock()

	if len(pending) == 0 {
		return
	}

	err := b.p.settle(ctx, pending)

	b.mu.Lock()
	defer b.mu.Unlock()

	if err != nil {
		err = fmt.Errorf("%w: %w", ErrSettle, err)
		b.errs = append(b.errs, err)
		b.p.config.logger.ErrorContext(ctx, "gacha: settle failed", slog.Int("entries", len(pending)),
			slog.Any("error", err))

		return
	}

	for i := range b.settled {
		b.settled[i] = true
	}
}

// settle settles outcomes in one call, detached from ctx's cancellation so that a shutdown does not lose them.
func (p *Poller[T, R]) settle(ctx context.Context, outcomes []gacha.Outcome) error {
	settleCtx, cancel := p.sourceContext(ctx)
	defer cancel()

	return p.source.Settle(settleCtx, outcomes...)
}

// unsettled returns the IDs of the entries that are not settled yet.
func (b *batch[T, R]) unsettled() []uuid.UUID {
	b.mu.Lock()
	defer b.mu.Unlock()

	ids := make([]uuid.UUID, 0, len(b.entries))

	for i := range b.entries {
		if !b.settled[i] {
			ids = append(ids, b.entries[i].ID)
		}
	}

	return ids
}

func (b *batch[T, R]) result() Result {
	res := Result{Claimed: len(b.entries)}

	for i := range b.outcomes {
		switch b.outcomes[i].Status {
		case gacha.StatusCompleted:
			res.Completed++
		case gacha.StatusPending, gacha.StatusReleased:
			res.Retried++
		case gacha.StatusFailed:
			res.Failed++
		case gacha.StatusProcessing:
			// Never decided.
		}
	}

	b.mu.Lock()
	res.Err = errors.Join(b.errs...)
	b.mu.Unlock()

	return res
}

// heartbeat renews the leases of the batch's unsettled entries until the returned function is called, when the
// source supports it.
func (p *Poller[T, R]) heartbeat(ctx context.Context, b *batch[T, R]) (stop func()) {
	if p.extender == nil || p.config.heartbeat <= 0 {
		return func() {}
	}

	// Keep beating during a shutdown: the handlers may still be returning.
	ctx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	done := make(chan struct{})

	go func() {
		defer close(done)

		ticker := time.NewTicker(p.config.heartbeat)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}

			ids := b.unsettled()
			if len(ids) == 0 {
				continue
			}

			p.extend(ctx, ids)
		}
	}()

	return func() {
		cancel()
		<-done
	}
}

// extend renews the leases of ids, bounded by the source timeout so that a stuck call cannot hold up the batch.
func (p *Poller[T, R]) extend(ctx context.Context, ids []uuid.UUID) {
	extendCtx, cancel := context.WithTimeout(ctx, p.config.sourceTimeout)
	defer cancel()

	if err := p.extender.Extend(extendCtx, p.config.lease, ids...); err != nil && ctx.Err() == nil {
		p.config.logger.ErrorContext(ctx, "gacha: heartbeat failed",
			slog.Int("entries", len(ids)),
			slog.Any("error", fmt.Errorf("%w: %w", ErrExtend, err)),
		)
	}
}
