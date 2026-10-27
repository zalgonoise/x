// Package mysql provides a gacha.Repository backed by MySQL (8.0 or later) or MariaDB (10.6 or later), the versions
// that support SKIP LOCKED, through database/sql and github.com/go-sql-driver/mysql.
//
// Each stream has its own outbox table and dead-letter table, created by Migrate (or by applying Schema with a
// migration tool). MySQL has no UPDATE ... RETURNING, so a claim is a short transaction: a SELECT ... FOR UPDATE SKIP
// LOCKED, then an UPDATE of the selected rows, committed before the handlers run. The repository's transactions run
// at READ COMMITTED, so their locking reads take no gap locks that would block concurrent writes. Single statements
// (most settles, extends and writes) autocommit at the session's isolation level instead.
//
// All times are the database's UTC_TIMESTAMP(6), so leases and retries are immune to clock skew between hosts and to
// session time zones.
//
// Recommended DSN parameters:
//
//   - interpolateParams=true: without it, every statement with arguments costs extra round-trips to prepare and
//     close it.
//   - transaction_isolation=%27READ-COMMITTED%27: makes the autocommitted statements READ COMMITTED too, so that a
//     settle of an entry that no longer exists takes no gap lock (InnoDB's default, REPEATABLE READ, does).
//
// parseTime is not required.
package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"uuid"

	"github.com/zalgonoise/cfg"

	"github.com/zalgonoise/x/gacha"
)

// execer runs statements, on a *sql.DB or a *sql.Tx.
type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// Repository is a MySQL outbox. It implements gacha.Repository, gacha.Extender, gacha.Redriver, gacha.Stater and
// gacha.Viewer. It is safe for concurrent use.
type Repository struct {
	db               *sql.DB
	table            string
	dlq              string
	deleteOnComplete bool
	schema           []string
	claimQuery       string
	statsQuery       string
}

// New creates a Repository over db, for the outbox table named table. The name is required, so that streams of
// different item types never share a table by accident, and may be qualified with a database name ("db.table"). New
// does not touch the database: call Migrate, or apply Schema, to create the tables.
func New(db *sql.DB, table string, opts ...cfg.Option[Config]) (*Repository, error) {
	config := cfg.Set(defaultConfig(), opts...)

	out, dlq, err := tables(table, config.dlqTable)
	if err != nil {
		return nil, err
	}

	return &Repository{
		db:               db,
		table:            out,
		dlq:              dlq,
		deleteOnComplete: config.deleteOnComplete,
		schema:           schemaStatements(out, dlq),
		claimQuery: "SELECT id, data, attempts, created_at FROM " + out +
			" WHERE claim_at <= " + now + " ORDER BY claim_at, id LIMIT ? FOR UPDATE SKIP LOCKED",
		statsQuery: "SELECT COALESCE(SUM(status = 'pending'), 0), COALESCE(SUM(status = 'processing'), 0), " +
			"COALESCE(SUM(status = 'completed'), 0), (SELECT COUNT(*) FROM " + dlq + ") FROM " + out,
	}, nil
}

// Migrate creates the outbox table and the dead-letter table, with their indexes, if they do not exist.
func (r *Repository) Migrate(ctx context.Context) error {
	for _, query := range r.schema {
		if _, err := r.db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("%s: migrate: %w", domain, err)
		}
	}

	return nil
}

// Write stores data as new pending entries, available for claims immediately. Large batches are split into several
// INSERT statements within one transaction, so the write stays atomic.
func (r *Repository) Write(ctx context.Context, data ...[]byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	statements := r.inserts(data)

	return r.exec(ctx, "write", statements)
}

// Tx returns a sink that writes to the outbox within tx, so the entries commit or roll back together with the
// caller's business data.
func (r *Repository) Tx(tx *sql.Tx) gacha.Sink[[]byte] {
	return gacha.SinkFunc[[]byte](func(ctx context.Context, data ...[]byte) error {
		if err := ctx.Err(); err != nil {
			return err
		}

		if err := writeTx(ctx, tx, r.inserts(data)); err != nil {
			return fmt.Errorf("%s: write: %w", domain, err)
		}

		return nil
	})
}

// writeTx runs the INSERT statements of a write within the caller's transaction. Several statements run under a
// savepoint, so that a failure rolls back the whole write and leaves the transaction usable.
func writeTx(ctx context.Context, tx *sql.Tx, statements []statement) error {
	if len(statements) <= 1 {
		return execAll(ctx, tx, statements)
	}

	if _, err := tx.ExecContext(ctx, "SAVEPOINT gacha_write"); err != nil {
		return err
	}

	if err := execAll(ctx, tx, statements); err != nil {
		_, rollbackErr := tx.ExecContext(context.WithoutCancel(ctx), "ROLLBACK TO SAVEPOINT gacha_write")

		return errors.Join(err, rollbackErr)
	}

	_, err := tx.ExecContext(ctx, "RELEASE SAVEPOINT gacha_write")

	return err
}

// inserts returns the INSERT statements writing data, split by row count and size.
func (r *Repository) inserts(data [][]byte) []statement {
	var (
		statements []statement
		from, size int
	)

	for i := range data {
		if i > from && (i-from == maxRows || size+len(data[i]) > maxBytes) {
			statements = append(statements, r.insert(data[from:i]))
			from, size = i, 0
		}

		size += len(data[i])
	}

	if from < len(data) {
		statements = append(statements, r.insert(data[from:]))
	}

	return statements
}

func (r *Repository) insert(data [][]byte) statement {
	args := make([]any, 0, insertArgs*len(data))

	for i := range data {
		id := uuid.NewV7()
		value := data[i]

		if value == nil {
			value = []byte{}
		}

		args = append(args, id[:], value)
	}

	row := "(?, ?, " + now + ", " + now + ", " + now + ")"
	rows := row + strings.Repeat(", "+row, len(data)-1)

	return statement{
		query: "INSERT INTO " + r.table + " (id, data, created_at, updated_at, available_at) VALUES " + rows,
		args:  args,
	}
}

// Claim leases up to limit claimable entries for the lease duration, as described by gacha.Source.
func (r *Repository) Claim(ctx context.Context, limit int, lease time.Duration) ([]gacha.Entry[[]byte], error) {
	entries, err := r.claim(ctx, limit, lease, statement{query: r.claimQuery, args: []any{limit}},
		func(n int) string {
			return "UPDATE " + r.table + " SET status = 'processing', attempts = attempts + 1, updated_at = " + now +
				", available_at = DATE_ADD(" + now + ", INTERVAL ? MICROSECOND) WHERE id IN (" + placeholders(n) + ")"
		})
	if err != nil {
		return nil, err
	}

	// The SELECT read the attempt counts before the UPDATE incremented them.
	for i := range entries {
		entries[i].Attempts++
	}

	return entries, nil
}

// claim runs a claim transaction: the locking read sel, then the UPDATE built by update for the selected IDs, with
// the lease as its first argument.
func (r *Repository) claim(
	ctx context.Context, limit int, lease time.Duration, sel statement, update func(n int) string,
) ([]gacha.Entry[[]byte], error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if lease <= 0 {
		return nil, gacha.ErrInvalidLease
	}

	if limit <= 0 {
		return nil, nil
	}

	var entries []gacha.Entry[[]byte]

	err := r.inTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, sel.query, sel.args...)
		if err != nil {
			return err
		}

		if entries, err = scanEntries(rows, min(limit, maxRows)); err != nil {
			return err
		}

		for _, span := range spans(len(entries)) {
			args := make([]any, 1, span[1]-span[0]+1)
			args[0] = leaseMicros(lease)

			for i := span[0]; i < span[1]; i++ {
				args = append(args, entries[i].ID[:])
			}

			if _, err = tx.ExecContext(ctx, update(span[1]-span[0]), args...); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%s: claim: %w", domain, err)
	}

	return entries, nil
}

// Settle applies the outcomes atomically, as described by gacha.Source: a single statement, or a transaction when
// the outcomes need several. When several outcomes share an ID, the first one wins.
func (r *Repository) Settle(ctx context.Context, outcomes ...gacha.Outcome) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := validate(outcomes); err != nil {
		return err
	}

	completed, retried, failed, released := group(dedupe(outcomes))

	var statements []statement

	for _, span := range spans(len(completed)) {
		statements = append(statements, r.settleCompleted(completed[span[0]:span[1]]))
	}

	for _, span := range spans(len(retried)) {
		statements = append(statements, r.settleRetried(retried[span[0]:span[1]]))
	}

	for _, span := range spans(len(failed)) {
		statements = append(statements, r.settleFailed(failed[span[0]:span[1]])...)
	}

	for _, span := range spans(len(released)) {
		statements = append(statements, r.settleReleased(released[span[0]:span[1]]))
	}

	return r.exec(ctx, "settle", statements)
}

// fenced returns the derived table of outcomes joined to the outbox, matching rows with the attempt count of the claim
// being settled, with the retry delays when retry is set.
func fenced(outcomes []gacha.Outcome, retry bool) (join string, args []any) {
	row, perRow := outcomeRow, outcomeArgs
	if retry {
		row, perRow = retryRow, retryArgs
	}

	args = make([]any, 0, len(outcomes)*perRow)

	for i := range outcomes {
		args = append(args, outcomes[i].ID[:], outcomes[i].Attempts, outcomes[i].Err)

		if retry {
			args = append(args, interval(outcomes[i].RetryIn))
		}
	}

	return " AS t JOIN " + values(row, len(outcomes)) + " AS i ON t.id = i.id AND t.attempts = i.attempts", args
}

const lastError = "t.last_error = COALESCE(NULLIF(i.err, ''), t.last_error)"

func (r *Repository) settleCompleted(outcomes []gacha.Outcome) statement {
	join, args := fenced(outcomes, false)

	if r.deleteOnComplete {
		return statement{query: "DELETE t FROM " + r.table + join + " WHERE t.status = 'processing'", args: args}
	}

	return statement{
		query: "UPDATE " + r.table + join + " SET t.status = 'completed', t.updated_at = " + now +
			", t.available_at = " + now + ", " + lastError + " WHERE t.status = 'processing'",
		args: args,
	}
}

func (r *Repository) settleRetried(outcomes []gacha.Outcome) statement {
	join, args := fenced(outcomes, true)

	return statement{
		query: "UPDATE " + r.table + join + " SET t.status = 'pending', t.updated_at = " + now +
			", t.available_at = DATE_ADD(" + now + ", INTERVAL i.retry_us MICROSECOND), " + lastError +
			" WHERE t.status = 'processing'",
		args: args,
	}
}

// settleReleased hands outcomes back as pending, without counting their claim as an attempt.
func (r *Repository) settleReleased(outcomes []gacha.Outcome) statement {
	join, args := fenced(outcomes, false)

	return statement{
		query: "UPDATE " + r.table + join + " SET t.status = 'pending', t.attempts = t.attempts - 1, t.updated_at = " +
			now + ", t.available_at = " + now + ", " + lastError + " WHERE t.status = 'processing'",
		args: args,
	}
}

// settleFailed returns the statements moving outcomes to the dead-letter table. The rows are locked first: at READ
// COMMITTED, INSERT ... SELECT reads without locking, and a concurrent claim must not change them before the DELETE.
func (r *Repository) settleFailed(outcomes []gacha.Outcome) []statement {
	join, args := fenced(outcomes, false)

	ids := make([]any, len(outcomes))
	for i := range outcomes {
		ids[i] = outcomes[i].ID[:]
	}

	return []statement{
		{
			query: "SELECT id FROM " + r.table + " WHERE id IN (" + placeholders(len(ids)) + ") FOR UPDATE",
			args:  ids,
		},
		{
			query: "INSERT INTO " + r.dlq + " (id, data, attempts, last_error, created_at, failed_at, available_at) " +
				"SELECT t.id, t.data, t.attempts, COALESCE(NULLIF(i.err, ''), t.last_error), t.created_at, " +
				now + ", " + now + " FROM " + r.table + join + " WHERE t.status = 'processing'",
			args: args,
		},
		{
			query: "DELETE t FROM " + r.table + join + " WHERE t.status = 'processing'",
			args:  args,
		},
	}
}

// Extend renews the leases of the given processing entries, as described by gacha.Extender.
func (r *Repository) Extend(ctx context.Context, lease time.Duration, ids ...uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if lease <= 0 {
		return gacha.ErrInvalidLease
	}

	var statements []statement

	for _, span := range spans(len(ids)) {
		chunk := ids[span[0]:span[1]]

		statements = append(statements, statement{
			query: "UPDATE " + r.table + " SET available_at = DATE_ADD(" + now + ", INTERVAL ? MICROSECOND), " +
				"updated_at = " + now + " WHERE status = 'processing' AND id IN (" + placeholders(len(chunk)) + ")",
			args: idArgs([]any{leaseMicros(lease)}, chunk),
		})
	}

	return r.exec(ctx, "extend", statements)
}

// Redrive moves the given dead letters back to the outbox in one transaction, as described by gacha.Redriver.
func (r *Repository) Redrive(ctx context.Context, ids ...uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var statements []statement

	for _, span := range spans(len(ids)) {
		chunk := ids[span[0]:span[1]]
		in := " WHERE id IN (" + placeholders(len(chunk)) + ")"
		args := idArgs(nil, chunk)

		statements = append(statements,
			statement{query: "SELECT id FROM " + r.dlq + in + " FOR UPDATE", args: args},
			statement{
				query: "INSERT INTO " + r.table +
					" (id, data, status, attempts, last_error, created_at, updated_at, available_at) " +
					"SELECT id, data, 'pending', 0, last_error, created_at, " + now + ", " + now +
					" FROM " + r.dlq + in,
				args: args,
			},
			statement{query: "DELETE FROM " + r.dlq + in, args: args},
		)
	}

	return r.exec(ctx, "redrive", statements)
}

// Stats returns the number of entries in each state. It counts every row of both tables, so it suits monitoring
// at a modest rate rather than hot paths.
func (r *Repository) Stats(ctx context.Context) (gacha.Stats, error) {
	if err := ctx.Err(); err != nil {
		return gacha.Stats{}, err
	}

	var stats gacha.Stats

	err := r.db.QueryRowContext(ctx, r.statsQuery).
		Scan(&stats.Pending, &stats.Processing, &stats.Completed, &stats.DeadLetters)
	if err != nil {
		return gacha.Stats{}, fmt.Errorf("%s: stats: %w", domain, err)
	}

	return stats, nil
}

// exec runs statements on the database: a single statement on its own, several in one transaction.
func (r *Repository) exec(ctx context.Context, op string, statements []statement) error {
	var err error

	switch len(statements) {
	case 0:
		return nil
	case 1:
		err = execAll(ctx, r.db, statements)
	default:
		err = r.inTx(ctx, func(tx *sql.Tx) error {
			return execAll(ctx, tx, statements)
		})
	}

	if err != nil {
		return fmt.Errorf("%s: %s: %w", domain, op, err)
	}

	return nil
}

func execAll(ctx context.Context, db execer, statements []statement) error {
	for i := range statements {
		if _, err := db.ExecContext(ctx, statements[i].query, statements[i].args...); err != nil {
			return err
		}
	}

	return nil
}

// inTx runs fn in a READ COMMITTED transaction, committing it if fn succeeds.
func (r *Repository) inTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}

	if err = fn(tx); err != nil {
		_ = tx.Rollback()

		return err
	}

	return tx.Commit()
}

func validate(outcomes []gacha.Outcome) error {
	for i := range outcomes {
		if err := outcomes[i].Validate(); err != nil {
			return err
		}
	}

	return nil
}

// dedupe returns the outcomes without repeated IDs, keeping the first outcome of each ID.
func dedupe(outcomes []gacha.Outcome) []gacha.Outcome {
	if len(outcomes) <= 1 {
		return outcomes
	}

	seen := make(map[uuid.UUID]struct{}, len(outcomes))
	out := outcomes[:0:0]

	for i := range outcomes {
		if _, ok := seen[outcomes[i].ID]; ok {
			continue
		}

		seen[outcomes[i].ID] = struct{}{}
		out = append(out, outcomes[i])
	}

	return out
}

// group splits outcomes by status.
func group(outcomes []gacha.Outcome) (completed, retried, failed, released []gacha.Outcome) {
	for i := range outcomes {
		switch outcomes[i].Status {
		case gacha.StatusCompleted:
			completed = append(completed, outcomes[i])
		case gacha.StatusPending:
			retried = append(retried, outcomes[i])
		case gacha.StatusFailed:
			failed = append(failed, outcomes[i])
		case gacha.StatusReleased:
			released = append(released, outcomes[i])
		case gacha.StatusProcessing:
			// Rejected by validate.
		}
	}

	return completed, retried, failed, released
}
