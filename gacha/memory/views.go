package memory

import (
	"container/list"
	"context"
	"time"
	"uuid"

	"github.com/zalgonoise/x/gacha"
)

// view is a gacha.Source over the completed records or the dead letters of a Repository.
type view[R any] struct {
	repo      *Repository[R]
	retention time.Duration
	dead      bool
}

// Completed returns a view over the completed entries last updated at least retention ago, as described by
// gacha.Viewer. It is always empty when the repository was created with WithDeleteOnComplete.
func (r *Repository[R]) Completed(retention time.Duration) gacha.Source[R] {
	return view[R]{repo: r, retention: retention}
}

// DeadLetters returns a view over the dead letters that failed at least retention ago, as described by gacha.Viewer.
func (r *Repository[R]) DeadLetters(retention time.Duration) gacha.Source[R] {
	return view[R]{repo: r, retention: retention, dead: true}
}

// records returns the view's list and the lookup for its IDs. The caller must hold the repository's lock.
func (v view[R]) records() (*list.List, func(id uuid.UUID) (*record[R], bool)) {
	if v.dead {
		return v.repo.deadList, func(id uuid.UUID) (*record[R], bool) {
			rec, ok := v.repo.dead[id]

			return rec, ok
		}
	}

	return v.repo.completed, func(id uuid.UUID) (*record[R], bool) {
		rec, ok := v.repo.entries[id]

		return rec, ok && rec.status == gacha.StatusCompleted
	}
}

// Claim leases up to limit of the view's entries, oldest first. It does not change their attempt counts.
func (v view[R]) Claim(ctx context.Context, limit int, lease time.Duration) ([]gacha.Entry[R], error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if lease <= 0 {
		return nil, gacha.ErrInvalidLease
	}

	if limit <= 0 {
		return nil, nil
	}

	v.repo.mu.Lock()
	defer v.repo.mu.Unlock()

	now := time.Now()
	expiry := now.Add(lease)
	cutoff := now.Add(-v.retention)
	records, _ := v.records()

	var entries []gacha.Entry[R]

	// Lists are ordered by updatedAt, which never changes once a record joins them.
	for elem := records.Front(); elem != nil && len(entries) < limit; elem = elem.Next() {
		rec, _ := elem.Value.(*record[R])
		if rec.updatedAt.After(cutoff) {
			break
		}

		if rec.availableAt.After(now) {
			continue
		}

		rec.availableAt = expiry
		entries = append(entries, rec.entry())
	}

	return entries, nil
}

// Settle applies the outcomes atomically: StatusCompleted deletes the entry, StatusPending makes it claimable again
// after RetryIn, and StatusFailed leaves it leased.
func (v view[R]) Settle(ctx context.Context, outcomes ...gacha.Outcome) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := validate(outcomes); err != nil {
		return err
	}

	v.repo.mu.Lock()
	defer v.repo.mu.Unlock()

	now := time.Now()
	records, lookup := v.records()

	for i := range outcomes {
		rec, ok := lookup(outcomes[i].ID)
		if !ok {
			continue
		}

		switch outcomes[i].Status {
		case gacha.StatusCompleted:
			records.Remove(rec.elem)
			rec.elem = nil

			if v.dead {
				delete(v.repo.dead, rec.id)

				continue
			}

			delete(v.repo.entries, rec.id)

		case gacha.StatusPending:
			rec.availableAt = now.Add(max(outcomes[i].RetryIn, 0))

		case gacha.StatusReleased:
			rec.availableAt = now

		case gacha.StatusFailed, gacha.StatusProcessing:
			// A failed entry stays leased until its lease expires. Processing is rejected by validate.
		}
	}

	return nil
}
