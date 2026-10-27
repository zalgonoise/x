package mysql

import (
	"context"
	"time"

	"github.com/zalgonoise/x/gacha"
)

// view is a gacha.Source over the completed rows of the outbox, or over the dead-letter table.
type view struct {
	repo  *Repository
	table string
	// since is the column the retention applies to, and the view's order. completed_at is NULL for rows that are not
	// completed, so the Completed view's claims never see them.
	since string
	// completedOnly restricts settles to completed rows, since the outbox holds rows in other states too.
	completedOnly bool
	retention     time.Duration
}

// Completed returns a view over the completed entries last updated at least retention ago, as described by
// gacha.Viewer. It is always empty when the repository was created with WithDeleteOnComplete.
func (r *Repository) Completed(retention time.Duration) gacha.Source[[]byte] {
	return view{repo: r, table: r.table, since: "completed_at", completedOnly: true, retention: retention}
}

// DeadLetters returns a view over the dead letters that failed at least retention ago, as described by gacha.Viewer.
func (r *Repository) DeadLetters(retention time.Duration) gacha.Source[[]byte] {
	return view{repo: r, table: r.dlq, since: "failed_at", retention: retention}
}

// Claim leases up to limit of the view's entries, oldest first. It does not change their attempt counts.
func (v view) Claim(ctx context.Context, limit int, lease time.Duration) ([]gacha.Entry[[]byte], error) {
	sel := statement{
		query: "SELECT id, data, attempts, created_at FROM " + v.table +
			" WHERE " + v.since + " <= DATE_SUB(" + now + ", INTERVAL ? MICROSECOND) AND available_at <= " + now +
			" ORDER BY " + v.since + ", id LIMIT ? FOR UPDATE SKIP LOCKED",
		args: []any{interval(v.retention), limit},
	}

	return v.repo.claim(ctx, limit, lease, sel, func(n int) string {
		return "UPDATE " + v.table + " SET available_at = DATE_ADD(" + now + ", INTERVAL ? MICROSECOND)" +
			" WHERE id IN (" + placeholders(n) + ")"
	})
}

// Settle applies the outcomes atomically: StatusCompleted deletes the entry, StatusPending makes it claimable again
// after RetryIn, and StatusFailed leaves it leased.
func (v view) Settle(ctx context.Context, outcomes ...gacha.Outcome) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := validate(outcomes); err != nil {
		return err
	}

	var deleteFilter, retryFilter string

	if v.completedOnly {
		deleteFilter = " AND status = 'completed'"
		retryFilter = " WHERE t.status = 'completed'"
	}

	completed, retried, _, released := group(dedupe(outcomes))

	// A released view entry is claimable again right away.
	for i := range released {
		released[i].RetryIn = 0
	}

	retried = append(retried, released...)

	var statements []statement

	for _, span := range spans(len(completed)) {
		chunk := completed[span[0]:span[1]]
		args := make([]any, len(chunk))

		for i := range chunk {
			args[i] = chunk[i].ID[:]
		}

		statements = append(statements, statement{
			query: "DELETE FROM " + v.table + " WHERE id IN (" + placeholders(len(chunk)) + ")" + deleteFilter,
			args:  args,
		})
	}

	for _, span := range spans(len(retried)) {
		chunk := retried[span[0]:span[1]]
		args := make([]any, 0, delayArgs*len(chunk))

		for i := range chunk {
			args = append(args, chunk[i].ID[:], interval(chunk[i].RetryIn))
		}

		statements = append(statements, statement{
			query: "UPDATE " + v.table + " AS t JOIN " + values(delayRow, len(chunk)) +
				" AS i ON t.id = i.id SET t.available_at = DATE_ADD(" + now + ", INTERVAL i.retry_us MICROSECOND)" +
				retryFilter,
			args: args,
		})
	}

	return v.repo.exec(ctx, "view settle", statements)
}
