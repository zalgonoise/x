// Package memory provides an in-memory gacha.Repository, generic over the raw type R.
//
// It has the same semantics as the SQL repositories (leases, retries, fenced settles, the dead-letter table, the
// Completed and DeadLetters views), which makes it the reference implementation for tests and development. Entries
// do not survive the process, so it is not meant for production outboxes.
//
// Raw values are stored as given and returned as stored. With a reference type such as []byte, neither the writer
// nor the handler should modify a value after handing it over: the gacha.Bytes codec copies on marshal for this
// reason. With gacha.Identity, the repository stores items without marshaling them at all.
//
// The repository uses the process clock (time.Now), so tests can control it with testing/synctest.
package memory

import (
	"container/heap"
	"container/list"
	"context"
	"sync"
	"time"
	"uuid"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/gacha"
)

// Repository is an in-memory outbox. It implements gacha.Repository, gacha.Extender, gacha.Redriver, gacha.Stater
// and gacha.Viewer. It is safe for concurrent use.
//
// The zero value is not usable: create repositories with New.
type Repository[R any] struct {
	mu sync.Mutex

	deleteOnComplete bool

	// entries holds the outbox: pending, processing and completed records.
	entries map[uuid.UUID]*record[R]
	// queue holds the pending and processing records, for claims.
	queue queue[R]
	// processing counts the processing records in queue.
	processing int
	// completed holds the completed records, by completion time.
	completed *list.List

	// dead holds the dead-letter table.
	dead map[uuid.UUID]*record[R]
	// deadList holds the dead letters, by failure time.
	deadList *list.List
}

// New creates an empty in-memory Repository.
func New[R any](opts ...cfg.Option[Config]) *Repository[R] {
	config := cfg.Set(defaultConfig(), opts...)

	return &Repository[R]{
		deleteOnComplete: config.deleteOnComplete,
		entries:          make(map[uuid.UUID]*record[R]),
		completed:        list.New(),
		dead:             make(map[uuid.UUID]*record[R]),
		deadList:         list.New(),
	}
}

// Write stores data as new pending entries, available for claims immediately.
func (r *Repository[R]) Write(ctx context.Context, data ...R) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	for i := range data {
		rec := &record[R]{
			id:          uuid.NewV7(),
			data:        data[i],
			status:      gacha.StatusPending,
			createdAt:   now,
			updatedAt:   now,
			availableAt: now,
			index:       notQueued,
		}

		r.entries[rec.id] = rec
		heap.Push(&r.queue, rec)
	}

	return nil
}

// Claim leases up to limit claimable entries for the lease duration, as described by gacha.Source.
func (r *Repository[R]) Claim(ctx context.Context, limit int, lease time.Duration) ([]gacha.Entry[R], error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if lease <= 0 {
		return nil, gacha.ErrInvalidLease
	}

	if limit <= 0 {
		return nil, nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	expiry := now.Add(lease)
	entries := make([]gacha.Entry[R], 0, min(limit, len(r.queue)))

	// Claimed records move behind the claim time, so the queue's head is always the next claimable record.
	for len(entries) < limit && len(r.queue) > 0 {
		rec := r.queue[0]
		if rec.availableAt.After(now) {
			break
		}

		if rec.status == gacha.StatusPending {
			rec.status = gacha.StatusProcessing
			r.processing++
		}

		rec.attempts++
		rec.updatedAt = now
		rec.availableAt = expiry

		heap.Fix(&r.queue, 0)

		entries = append(entries, rec.entry())
	}

	return entries, nil
}

// Settle applies the outcomes atomically, as described by gacha.Source.
func (r *Repository[R]) Settle(ctx context.Context, outcomes ...gacha.Outcome) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := validate(outcomes); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	for i := range outcomes {
		rec, ok := r.entries[outcomes[i].ID]
		if !ok || rec.status != gacha.StatusProcessing || rec.attempts != outcomes[i].Attempts {
			continue
		}

		r.settle(rec, &outcomes[i], now)
	}

	return nil
}

func (r *Repository[R]) settle(rec *record[R], outcome *gacha.Outcome, now time.Time) {
	if outcome.Err != "" {
		rec.lastError = outcome.Err
	}

	rec.updatedAt = now

	switch outcome.Status {
	case gacha.StatusPending, gacha.StatusReleased:
		rec.status = gacha.StatusPending
		rec.availableAt = now.Add(max(outcome.RetryIn, 0))
		r.processing--

		if outcome.Status == gacha.StatusReleased {
			// The claim does not count as an attempt.
			rec.attempts--
			rec.availableAt = now
		}

		heap.Fix(&r.queue, rec.index)

	case gacha.StatusCompleted:
		r.dequeue(rec)

		if r.deleteOnComplete {
			delete(r.entries, rec.id)

			return
		}

		rec.status = gacha.StatusCompleted
		rec.availableAt = now
		rec.elem = r.completed.PushBack(rec)

	case gacha.StatusFailed:
		r.dequeue(rec)
		delete(r.entries, rec.id)

		rec.status = gacha.StatusFailed
		rec.availableAt = now
		r.dead[rec.id] = rec
		rec.elem = r.deadList.PushBack(rec)

	case gacha.StatusProcessing:
		// Rejected by validate.
	}
}

// dequeue removes a processing record from the claim queue.
func (r *Repository[R]) dequeue(rec *record[R]) {
	heap.Remove(&r.queue, rec.index)
	r.processing--
}

// Extend renews the leases of the given processing entries, as described by gacha.Extender.
func (r *Repository[R]) Extend(ctx context.Context, lease time.Duration, ids ...uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if lease <= 0 {
		return gacha.ErrInvalidLease
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	expiry := now.Add(lease)

	for _, id := range ids {
		rec, ok := r.entries[id]
		if !ok || rec.status != gacha.StatusProcessing {
			continue
		}

		rec.updatedAt = now
		rec.availableAt = expiry
		heap.Fix(&r.queue, rec.index)
	}

	return nil
}

// Redrive moves the given dead letters back to the outbox, as described by gacha.Redriver.
func (r *Repository[R]) Redrive(ctx context.Context, ids ...uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	for _, id := range ids {
		rec, ok := r.dead[id]
		if !ok {
			continue
		}

		delete(r.dead, id)
		r.deadList.Remove(rec.elem)

		rec.elem = nil
		rec.status = gacha.StatusPending
		rec.attempts = 0
		rec.updatedAt = now
		rec.availableAt = now
		r.entries[id] = rec
		heap.Push(&r.queue, rec)
	}

	return nil
}

// Stats returns the number of entries in each state.
func (r *Repository[R]) Stats(ctx context.Context) (gacha.Stats, error) {
	if err := ctx.Err(); err != nil {
		return gacha.Stats{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	return gacha.Stats{
		Pending:     int64(len(r.queue) - r.processing),
		Processing:  int64(r.processing),
		Completed:   int64(r.completed.Len()),
		DeadLetters: int64(len(r.dead)),
	}, nil
}

func validate(outcomes []gacha.Outcome) error {
	for i := range outcomes {
		if err := outcomes[i].Validate(); err != nil {
			return err
		}
	}

	return nil
}
