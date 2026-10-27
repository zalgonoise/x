package gacha

import (
	"context"
	"time"
	"uuid"
)

// Sink persists a batch of raw items atomically: either every item is stored, or none is.
//
// Every repository implements Sink. SQL repositories also expose sinks bound to a caller's transaction, so the outbox
// write commits or rolls back together with the caller's business data.
//
// Write must not retain data after it returns. Calling Write without data is a no-op.
type Sink[R any] interface {
	Write(ctx context.Context, data ...R) error
}

// SinkFunc is a function type that implements Sink.
type SinkFunc[R any] func(ctx context.Context, data ...R) error

// Write calls fn(ctx, data...).
func (fn SinkFunc[R]) Write(ctx context.Context, data ...R) error {
	return fn(ctx, data...)
}

// Source claims entries for processing and settles their outcomes.
//
// Repositories implement Source over their claimable entries. Their Completed and DeadLetters views (see Viewer)
// implement it over completed and dead-lettered entries.
type Source[R any] interface {
	// Claim leases up to limit claimable entries for the lease duration, increments their attempt counts, and returns
	// them ordered by availability time, then ID (views order them by age instead, see Viewer). The claim is durable
	// before Claim returns, so the attempt is accounted for even if the process crashes afterwards. An entry whose
	// lease expires without being settled can be claimed again.
	//
	// A non-positive limit claims nothing. A non-positive lease returns ErrInvalidLease.
	Claim(ctx context.Context, limit int, lease time.Duration) ([]Entry[R], error)

	// Settle applies the outcomes atomically.
	//
	// Outcomes that do not match a claimed entry are ignored: unknown IDs, entries that were already settled, and
	// entries claimed again since (Outcome.Attempts differs from the current attempt count). An outcome whose status
	// is not StatusCompleted, StatusPending, StatusFailed or StatusReleased fails the whole call with
	// ErrInvalidStatus.
	Settle(ctx context.Context, outcomes ...Outcome) error
}

// Repository is the outbox: a Sink to write entries to and a Source to claim them from.
type Repository[R any] interface {
	Sink[R]
	Source[R]
}

// Writer marshals items and writes them to a Sink. It returns once the items are persisted.
type Writer[T any] interface {
	Write(ctx context.Context, items ...T) error
}

// Handler processes one item. The entry's ID is available through IDFromContext.
//
// A nil error completes the entry. Any other error is handled by the poller's retry policy, except for errors marked
// with Permanent, which move the entry to the dead-letter table immediately.
type Handler[T any] func(ctx context.Context, item T) error

// Extender is an optional Source capability that renews the leases of claimed entries. The poller uses it to keep
// the entries of long-running handlers from being claimed again.
type Extender interface {
	// Extend sets the leases of the given processing entries to expire after lease. Entries that are not processing
	// are ignored. A non-positive lease returns ErrInvalidLease.
	Extend(ctx context.Context, lease time.Duration, ids ...uuid.UUID) error
}

// Redriver is an optional Repository capability that moves dead-lettered entries back to the outbox.
type Redriver interface {
	// Redrive moves the given dead-lettered entries back to the outbox as pending, claimable entries with no
	// attempts. Their IDs, data and creation times are kept. IDs not in the dead-letter table are ignored.
	Redrive(ctx context.Context, ids ...uuid.UUID) error
}

// Stater is an optional Repository capability that reports the number of entries in each state.
type Stater interface {
	Stats(ctx context.Context) (Stats, error)
}

// Viewer is an optional Repository capability that exposes Sources over completed and dead-lettered entries, so
// they can be archived or purged by a poller, such as the one built by the janitor package.
//
// A view's Claim leases its rows like a regular claim, but it neither increments nor fences attempts: entries report
// their final attempt count. Settling a view entry as StatusCompleted deletes it, StatusPending makes it claimable
// again after Outcome.RetryIn, StatusReleased makes it claimable again right away, and StatusFailed leaves it leased
// until the lease expires.
type Viewer[R any] interface {
	// Completed returns a view over the completed entries last updated at least retention ago, oldest first.
	Completed(retention time.Duration) Source[R]
	// DeadLetters returns a view over the dead-lettered entries that failed at least retention ago, oldest first.
	DeadLetters(retention time.Duration) Source[R]
}
