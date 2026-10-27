package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zalgonoise/x/gacha"
)

// view is a gacha.Source over the completed rows of the outbox, or over the dead-letter table.
type view struct {
	repo      *Repository
	claim     string
	settle    string
	retention time.Duration
}

// Completed returns a view over the completed entries last updated at least retention ago, as described by
// gacha.Viewer. It is always empty when the repository was created with WithDeleteOnComplete.
func (r *Repository) Completed(retention time.Duration) gacha.Source[[]byte] {
	return view{repo: r, claim: r.q.completedClaim, settle: r.q.completedSettle, retention: max(retention, 0)}
}

// DeadLetters returns a view over the dead letters that failed at least retention ago, as described by gacha.Viewer.
func (r *Repository) DeadLetters(retention time.Duration) gacha.Source[[]byte] {
	return view{repo: r, claim: r.q.deadLettersClaim, settle: r.q.deadLettersSettle, retention: max(retention, 0)}
}

// Claim leases up to limit of the view's entries, oldest first. It does not change their attempt counts.
func (v view) Claim(ctx context.Context, limit int, lease time.Duration) ([]gacha.Entry[[]byte], error) {
	return v.repo.claim(ctx, v.claim, "view claim", limit, lease, v.retention)
}

// Settle applies the outcomes atomically in one statement: StatusCompleted deletes the entry, StatusPending makes it
// claimable again after RetryIn, and StatusFailed leaves it leased.
func (v view) Settle(ctx context.Context, outcomes ...gacha.Outcome) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := validate(outcomes); err != nil {
		return err
	}

	var (
		deleted   []pgtype.UUID
		retried   []pgtype.UUID
		retriedIn []pgtype.Interval
	)

	for _, outcome := range dedupe(outcomes) {
		switch outcome.Status {
		case gacha.StatusCompleted:
			deleted = append(deleted, pgUUID(outcome.ID))
		case gacha.StatusPending:
			retried = append(retried, pgUUID(outcome.ID))
			retriedIn = append(retriedIn, pgInterval(outcome.RetryIn))
		case gacha.StatusReleased:
			retried = append(retried, pgUUID(outcome.ID))
			retriedIn = append(retriedIn, pgInterval(0))
		case gacha.StatusFailed, gacha.StatusProcessing:
			// A failed entry stays leased until its lease expires. Processing is rejected by validate.
		}
	}

	if len(deleted) == 0 && len(retried) == 0 {
		return nil
	}

	if _, err := v.repo.db.Exec(ctx, v.settle, deleted, retried, retriedIn); err != nil {
		return fmt.Errorf("%s: view settle: %w", domain, err)
	}

	return nil
}
