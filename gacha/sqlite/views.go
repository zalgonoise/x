package sqlite

import (
	"context"
	"time"

	"github.com/zalgonoise/x/gacha"
)

// view is a gacha.Source over the completed rows of the outbox, or over the dead-letter table.
type view struct {
	repo  *Repository
	table string
	// since is the column the retention applies to, and the view's order.
	since string
	// completedOnly restricts the view to completed rows, since the outbox holds rows in other states too.
	completedOnly bool
	retention     time.Duration
}

// Completed returns a view over the completed entries last updated at least retention ago, as described by
// gacha.Viewer. It is always empty when the repository was created with WithDeleteOnComplete.
func (r *Repository) Completed(retention time.Duration) gacha.Source[[]byte] {
	return view{repo: r, table: r.table, since: "updated_at", completedOnly: true, retention: retention}
}

// DeadLetters returns a view over the dead letters that failed at least retention ago, as described by gacha.Viewer.
func (r *Repository) DeadLetters(retention time.Duration) gacha.Source[[]byte] {
	return view{repo: r, table: r.dlq, since: "failed_at", retention: retention}
}

// filter returns the condition restricting the view's rows, with the given column prefix.
func (v view) filter(prefix string) string {
	if !v.completedOnly {
		return ""
	}

	return " AND " + prefix + "status = 'completed'"
}

// claimQuery returns the read of the view's claims, whose arguments are the retention and the limit.
func (v view) claimQuery() string {
	return "SELECT id, data, attempts, created_at FROM " + v.table +
		" WHERE " + v.since + " <= " + now + " - ? AND available_at <= " + now + v.filter("") +
		" ORDER BY " + v.since + ", id LIMIT ?"
}

// Claim leases up to limit of the view's entries, oldest first. It does not change their attempt counts.
func (v view) Claim(ctx context.Context, limit int, lease time.Duration) ([]gacha.Entry[[]byte], error) {
	sel := statement{query: v.claimQuery(), args: []any{micros(v.retention), limit}}

	return v.repo.claim(ctx, limit, lease, sel, func(n int) string {
		return "UPDATE " + v.table + " SET available_at = " + now + " + ? WHERE id IN (" + placeholders(n) + ")"
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
			query: "DELETE FROM " + v.table + " WHERE id IN (" + placeholders(len(chunk)) + ")" + v.filter(""),
			args:  args,
		})
	}

	for _, span := range spans(len(retried)) {
		chunk := retried[span[0]:span[1]]
		args := make([]any, 0, delayArgs*len(chunk))

		for i := range chunk {
			args = append(args, chunk[i].ID[:], micros(chunk[i].RetryIn))
		}

		statements = append(statements, statement{
			query: with(delayColumns, delayRow, len(chunk)) + "UPDATE " + v.table + " AS t SET available_at = " +
				now + " + i.retry_us FROM i WHERE t.id = i.id" + v.filter("t."),
			args: args,
		})
	}

	return v.repo.exec(ctx, "view settle", statements)
}
