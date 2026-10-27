// Package sqlite provides a gacha.Repository backed by SQLite (3.35 or later), through database/sql. It works with
// any SQLite driver, and is tested with modernc.org/sqlite (pure Go, no cgo).
//
// Each stream has its own outbox table and dead-letter table, created by Migrate (or by applying Schema with a
// migration tool), with partial indexes for claims and the Completed view. Times are INTEGER microseconds since the
// Unix epoch, taken from the database clock (julianday('now'), with millisecond precision).
//
// # Concurrency
//
// SQLite allows one writer at a time, database-wide: claims never overlap, so they need no SKIP LOCKED. Operations
// spanning several statements (claims, dead-letter moves, redrives, large writes) run in BEGIN IMMEDIATE
// transactions, which take the write lock up front instead of failing to upgrade a read lock under contention.
//
// With several connections (a *sql.DB pool, several pollers, or several processes on the same file):
//
//   - Set a busy timeout, so that writers wait for the lock instead of failing with SQLITE_BUSY. With
//     modernc.org/sqlite: file:outbox.db?_pragma=busy_timeout(5000).
//   - Use WAL mode, so that readers do not block the writer and the writer does not block readers:
//     _pragma=journal_mode(WAL).
//   - Do not put the database on a network file system, whose locks SQLite cannot rely on.
//
// An in-memory database (":memory:") is private to each connection: use db.SetMaxOpenConns(1), or a shared-cache
// URI, so that every operation sees the same database.
//
// In WAL mode, keep the outbox and its dead-letter table in the same database file (see WithDeadLetterTable).
package sqlite

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"time"
	"uuid"

	"github.com/zalgonoise/cfg"

	"github.com/zalgonoise/x/gacha"
)

// execer runs statements, on a *sql.DB, a *sql.Conn or a *sql.Tx.
type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// Repository is a SQLite outbox. It implements gacha.Repository, gacha.Extender, gacha.Redriver, gacha.Stater and
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
// different item types never share a table by accident, and may be qualified with the name of an attached database
// ("aux.table"). New does not touch the database: call Migrate, or apply Schema, to create the tables.
func New(db *sql.DB, table string, opts ...cfg.Option[Config]) (*Repository, error) {
	config := cfg.Set(defaultConfig(), opts...)

	out, dlq, err := tables(table, config.dlqTable)
	if err != nil {
		return nil, err
	}

	return &Repository{
		db:               db,
		table:            out.ident,
		dlq:              dlq.ident,
		deleteOnComplete: config.deleteOnComplete,
		schema:           schemaStatements(out, dlq),
		claimQuery: "SELECT id, data, attempts, created_at FROM " + out.ident +
			" WHERE status IN ('pending', 'processing') AND available_at <= " + now +
			" ORDER BY available_at, id LIMIT ?",
		statsQuery: "SELECT COALESCE(SUM(status = 'pending'), 0), COALESCE(SUM(status = 'processing'), 0), " +
			"COALESCE(SUM(status = 'completed'), 0), (SELECT COUNT(*) FROM " + dlq.ident + ") FROM " + out.ident,
	}, nil
}

// Migrate creates the outbox table, the dead-letter table and their indexes if they do not exist, in one
// transaction.
func (r *Repository) Migrate(ctx context.Context) error {
	statements := make([]statement, len(r.schema))

	for i := range r.schema {
		statements[i] = statement{query: r.schema[i]}
	}

	if err := r.immediate(ctx, func(conn *sql.Conn) error {
		return execAll(ctx, conn, statements)
	}); err != nil {
		return fmt.Errorf("%s: migrate: %w", domain, err)
	}

	return nil
}

// Write stores data as new pending entries, available for claims immediately. Large batches are split into several
// INSERT statements within one transaction, so the write stays atomic.
func (r *Repository) Write(ctx context.Context, data ...[]byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return r.exec(ctx, "write", r.inserts(data))
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

// inserts returns the INSERT statements writing data, in chunks of maxRows rows.
func (r *Repository) inserts(data [][]byte) []statement {
	statements := make([]statement, 0, len(data)/maxRows+1)

	for _, span := range spans(len(data)) {
		chunk := data[span[0]:span[1]]
		args := make([]any, 0, insertArgs*len(chunk))

		for i := range chunk {
			id := uuid.NewV7()
			value := chunk[i]

			if value == nil {
				value = []byte{}
			}

			args = append(args, id[:], value)
		}

		statements = append(statements, statement{
			query: "INSERT INTO " + r.table + " (id, data, created_at, updated_at, available_at) " +
				"SELECT v.column1, v.column2, n.ts, n.ts, n.ts FROM (VALUES (?, ?)" +
				strings.Repeat(", (?, ?)", len(chunk)-1) + ") AS v, (SELECT " + now + " AS ts) AS n",
			args: args,
		})
	}

	return statements
}

// Claim leases up to limit claimable entries for the lease duration, as described by gacha.Source.
func (r *Repository) Claim(ctx context.Context, limit int, lease time.Duration) ([]gacha.Entry[[]byte], error) {
	entries, err := r.claim(ctx, limit, lease, statement{query: r.claimQuery, args: []any{limit}},
		func(n int) string {
			return "UPDATE " + r.table + " SET status = 'processing', attempts = attempts + 1, updated_at = " + now +
				", available_at = " + now + " + ? WHERE id IN (" + placeholders(n) + ")"
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

// claim runs a claim transaction: the read sel, then the UPDATE built by update for the selected IDs, with the lease
// as its first argument. The transaction holds the write lock throughout, so the rows cannot change in between.
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

	err := r.immediate(ctx, func(conn *sql.Conn) error {
		rows, err := conn.QueryContext(ctx, sel.query, sel.args...)
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

			if _, err = conn.ExecContext(ctx, update(span[1]-span[0]), args...); err != nil {
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

// fenced returns the CTE of outcomes, with the retry delays when retry is set, and its arguments. Statements join it
// to the outbox on the ID and the attempt count of the claim being settled.
func fenced(outcomes []gacha.Outcome, retry bool) (cte string, args []any) {
	columns, row, perRow := outcomeColumns, outcomeRow, outcomeArgs
	if retry {
		columns, row, perRow = retryColumns, retryRow, retryArgs
	}

	args = make([]any, 0, len(outcomes)*perRow)

	for i := range outcomes {
		args = append(args, outcomes[i].ID[:], outcomes[i].Attempts, outcomes[i].Err)

		if retry {
			args = append(args, micros(outcomes[i].RetryIn))
		}
	}

	return with(columns, row, len(outcomes)), args
}

const (
	lastError = "last_error = COALESCE(NULLIF(i.err, ''), t.last_error)"
	matches   = " WHERE t.id = i.id AND t.attempts = i.attempts AND t.status = 'processing'"
	exists    = " WHERE t.status = 'processing' AND EXISTS (SELECT 1 FROM i WHERE i.id = t.id AND i.attempts = t.attempts)"
)

func (r *Repository) settleCompleted(outcomes []gacha.Outcome) statement {
	cte, args := fenced(outcomes, false)

	if r.deleteOnComplete {
		return statement{query: cte + "DELETE FROM " + r.table + " AS t" + exists, args: args}
	}

	return statement{
		query: cte + "UPDATE " + r.table + " AS t SET status = 'completed', updated_at = " + now +
			", available_at = " + now + ", " + lastError + " FROM i" + matches,
		args: args,
	}
}

func (r *Repository) settleRetried(outcomes []gacha.Outcome) statement {
	cte, args := fenced(outcomes, true)

	return statement{
		query: cte + "UPDATE " + r.table + " AS t SET status = 'pending', updated_at = " + now +
			", available_at = " + now + " + i.retry_us, " + lastError + " FROM i" + matches,
		args: args,
	}
}

// settleReleased hands outcomes back as pending, without counting their claim as an attempt.
func (r *Repository) settleReleased(outcomes []gacha.Outcome) statement {
	cte, args := fenced(outcomes, false)

	return statement{
		query: cte + "UPDATE " + r.table + " AS t SET status = 'pending', attempts = t.attempts - 1, updated_at = " + now +
			", available_at = " + now + ", " + lastError + " FROM i" + matches,
		args: args,
	}
}

// settleFailed returns the statements moving outcomes to the dead-letter table. They run in one transaction, which
// holds the write lock, so the rows cannot change between the INSERT and the DELETE.
func (r *Repository) settleFailed(outcomes []gacha.Outcome) []statement {
	cte, args := fenced(outcomes, false)

	return []statement{
		{
			query: cte + "INSERT INTO " + r.dlq +
				" (id, data, attempts, last_error, created_at, failed_at, available_at) " +
				"SELECT t.id, t.data, t.attempts, COALESCE(NULLIF(i.err, ''), t.last_error), t.created_at, " +
				now + ", " + now + " FROM " + r.table + " AS t JOIN i ON t.id = i.id AND t.attempts = i.attempts" +
				" WHERE t.status = 'processing'",
			args: args,
		},
		{query: cte + "DELETE FROM " + r.table + " AS t" + exists, args: args},
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
			query: "UPDATE " + r.table + " SET available_at = " + now + " + ?, updated_at = " + now +
				" WHERE status = 'processing' AND id IN (" + placeholders(len(chunk)) + ")",
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

// exec runs statements on the database: a single statement on its own, several in one immediate transaction.
func (r *Repository) exec(ctx context.Context, op string, statements []statement) error {
	var err error

	switch len(statements) {
	case 0:
		return nil
	case 1:
		err = execAll(ctx, r.db, statements)
	default:
		err = r.immediate(ctx, func(conn *sql.Conn) error {
			return execAll(ctx, conn, statements)
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

// immediate runs fn in a BEGIN IMMEDIATE transaction on a dedicated connection, committing it if fn succeeds.
// database/sql only begins deferred transactions, whose read locks fail to upgrade under contention, so the
// transaction is driven by hand.
func (r *Repository) immediate(ctx context.Context, fn func(conn *sql.Conn) error) error {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = conn.Close()
	}()

	if _, err = conn.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		return err
	}

	if err = fn(conn); err == nil {
		if _, err = conn.ExecContext(ctx, "COMMIT"); err == nil {
			return nil
		}
	}

	if _, rollbackErr := conn.ExecContext(context.WithoutCancel(ctx), "ROLLBACK"); rollbackErr != nil {
		// Never hand a connection with an open transaction back to the pool.
		_ = conn.Raw(func(any) error {
			return driver.ErrBadConn
		})

		return errors.Join(err, rollbackErr)
	}

	return err
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
