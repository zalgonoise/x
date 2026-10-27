// Package poller provides the Poller, which claims entries from a gacha.Source, unmarshals them and hands each one
// to a gacha.Handler, along with the Triggers that decide when it polls and the retry backoffs.
//
// The failure policy lives in the poller, not in the handler:
//
//   - A handler error, a recovered panic or a handler timeout retries the entry after the retry backoff.
//   - Once an entry has failed WithMaxAttempts times, it is moved to the dead-letter table.
//   - Errors marked with gacha.Permanent, and entries that fail to unmarshal, are moved to the dead-letter table right
//     away.
//   - Entries claimed more times than the maximum, because handlers crashed the process before settling them, are
//     moved to the dead-letter table without running the handler.
//
// Each entry is settled as soon as its handler returns, unless WithBatchSettle is set. Settles are fenced by the
// claim's attempt count, so an entry whose lease expired and was claimed again is not settled twice.
package poller

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"uuid"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/gacha"
)

// Result is the outcome of one poll.
type Result struct {
	// Claimed is the number of entries claimed.
	Claimed int
	// Completed is the number of entries whose handler succeeded.
	Completed int
	// Retried is the number of entries handed back for a retry, including the ones released by a shutdown.
	Retried int
	// Failed is the number of entries moved to the dead-letter table.
	Failed int
	// Err reports a failed claim (ErrClaim) or failed settles (ErrSettle). The counts above are the poller's
	// verdicts, whether or not they were settled.
	Err error
}

// Poller claims entries from a gacha.Source in batches, unmarshals them, runs the handler on each one and settles
// the outcomes, following its retry policy.
type Poller[T, R any] struct {
	source   gacha.Source[R]
	extender gacha.Extender
	u        gacha.Unmarshaler[T, R]
	handler  gacha.Handler[T]
	config   Config
}

// New creates a Poller over source, unmarshaling entries with u and handling them with handler. All three must be
// non-nil. Heartbeats are enabled when source implements gacha.Extender: a Source that wraps a repository must
// implement it too to keep them.
func New[T, R any](
	source gacha.Source[R], u gacha.Unmarshaler[T, R], handler gacha.Handler[T], opts ...cfg.Option[Config],
) *Poller[T, R] {
	p := &Poller[T, R]{
		source:  source,
		u:       u,
		handler: handler,
		config:  cfg.Set(defaultConfig(), opts...).finish(),
	}

	p.extender, _ = source.(gacha.Extender)

	return p
}

// Run polls until ctx is done, waiting for the Trigger between polls (or not at all after a full batch, in drain
// mode). Claim and settle failures are reported to the observer and the logger, and do not stop Run.
//
// When ctx is done, the handlers of the current batch see their context canceled. Their outcomes are still settled.
// The entries they fail on, and the entries not started yet, are released (gacha.StatusReleased): handed back for an
// immediate retry, without counting the interrupted claim as an attempt. Run then returns nil.
//
// Run returns early only if the Trigger fails.
func (p *Poller[T, R]) Run(ctx context.Context) error {
	for !done(ctx) {
		res := p.poll(ctx)

		if done(ctx) || p.config.drain && res.Err == nil && res.Claimed >= p.config.batchSize {
			continue
		}

		if err := p.config.trigger.Wait(ctx, res); err != nil && !done(ctx) {
			return err
		}
	}

	return nil
}

func done(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

// Poll runs a single iteration: it claims a batch, handles it and settles it. It returns the poll's Result, and its
// Err as the error. It suits cron jobs and tests.
func (p *Poller[T, R]) Poll(ctx context.Context) (Result, error) {
	res := p.poll(ctx)

	return res, res.Err
}

func (p *Poller[T, R]) poll(ctx context.Context) Result {
	res := p.process(ctx)

	if res.Claimed > 0 {
		p.config.logger.DebugContext(ctx, "gacha: poll",
			slog.Int("claimed", res.Claimed),
			slog.Int("completed", res.Completed),
			slog.Int("retried", res.Retried),
			slog.Int("failed", res.Failed),
		)
	}

	if p.config.observer != nil {
		p.config.observer(res)
	}

	return res
}

func (p *Poller[T, R]) process(ctx context.Context) Result {
	if err := ctx.Err(); err != nil {
		return Result{Err: err}
	}

	// The claim runs detached from ctx's cancellation: a claim canceled mid-flight may still commit, leaving its
	// entries leased and unhandled. A claim that completes during a shutdown is released instead.
	claimCtx, cancel := p.sourceContext(ctx)
	entries, err := p.source.Claim(claimCtx, p.config.batchSize, p.config.lease)

	cancel()

	if err != nil {
		err = fmt.Errorf("%w: %w", ErrClaim, err)
		p.config.logger.ErrorContext(ctx, "gacha: claim failed", slog.Any("error", err))

		return Result{Err: err}
	}

	if len(entries) == 0 {
		return Result{}
	}

	b := newBatch(p, entries)

	// Heartbeats run until the whole batch is settled, panics included.
	stop := p.heartbeat(ctx, b)
	defer stop()

	b.handle(ctx)
	b.settleRemaining(ctx)

	return b.result()
}

// sourceContext returns a context for a call to the source: detached from ctx's cancellation, so that a shutdown
// does not lose claims or outcomes, and bounded by the source timeout.
func (p *Poller[T, R]) sourceContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), p.config.sourceTimeout)
}

// decide runs the handler on entry and returns the outcome dictated by the retry policy.
func (p *Poller[T, R]) decide(ctx context.Context, entry *gacha.Entry[R]) gacha.Outcome {
	outcome := gacha.Outcome{ID: entry.ID, Attempts: entry.Attempts}

	if p.config.maxAttempts > 0 && entry.Attempts > p.config.maxAttempts {
		return fail(outcome, ErrMaxAttempts)
	}

	item, err := p.unmarshal(ctx, entry)
	if err != nil {
		return fail(outcome, fmt.Errorf("%w: %w", ErrUnmarshal, err))
	}

	err = p.call(ctx, entry.ID, item)

	switch {
	case err == nil:
		outcome.Status = gacha.StatusCompleted

	case gacha.IsPermanent(err):
		return fail(outcome, err)

	case ctx.Err() != nil:
		// Shutting down: hand the entry back right away, without counting this attempt.
		outcome.Status = gacha.StatusReleased
		outcome.Err = err.Error()

	case p.config.maxAttempts > 0 && entry.Attempts >= p.config.maxAttempts:
		return fail(outcome, err)

	default:
		outcome.Status = gacha.StatusPending
		outcome.Err = err.Error()
		outcome.RetryIn = p.config.retryBackoff(entry.Attempts)
	}

	return outcome
}

// unmarshal decodes the entry's data, recovering panics: a panicking Unmarshaler fails the entry like an unmarshal
// error instead of crashing the process.
func (p *Poller[T, R]) unmarshal(ctx context.Context, entry *gacha.Entry[R]) (item T, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%w: %v", ErrPanic, r)

			p.config.logger.ErrorContext(ctx, "gacha: unmarshal panic",
				slog.String("id", entry.ID.String()),
				slog.Any("panic", r),
				slog.String("stack", string(debug.Stack())),
			)
		}
	}()

	return p.u.Unmarshal(entry.Data)
}

// call runs the handler with the entry's ID in its context, recovering panics.
func (p *Poller[T, R]) call(ctx context.Context, id uuid.UUID, item T) (err error) {
	ctx = gacha.ContextWithID(ctx, id)

	if p.config.handlerTimeout > 0 {
		var cancel context.CancelFunc

		ctx, cancel = context.WithTimeout(ctx, p.config.handlerTimeout)
		defer cancel()
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%w: %v", ErrPanic, r)

			p.config.logger.ErrorContext(ctx, "gacha: handler panic",
				slog.String("id", id.String()),
				slog.Any("panic", r),
				slog.String("stack", string(debug.Stack())),
			)
		}
	}()

	return p.handler(ctx, item)
}

func fail(outcome gacha.Outcome, err error) gacha.Outcome {
	outcome.Status = gacha.StatusFailed
	outcome.Err = err.Error()

	return outcome
}
