// Package postgres provides a gacha.Repository backed by PostgreSQL, using github.com/jackc/pgx/v5.
//
// Each stream has its own outbox table and dead-letter table, created by Migrate (or by applying Schema with a
// migration tool). Every operation is a single statement, so it is atomic and costs one round-trip:
//
//   - Write inserts a whole batch with unnest, in any size.
//   - Claim leases rows with UPDATE ... FROM (SELECT ... FOR UPDATE SKIP LOCKED) RETURNING, so concurrent pollers
//     never claim the same row, and committed claims hold no locks or connections while handlers run.
//   - Settle applies any mix of outcomes, moving dead letters with a DELETE ... RETURNING feeding an INSERT.
//
// Times are the database's (now() plus interval parameters), so leases and retries are immune to clock skew between
// hosts. Tx binds writes to a caller's transaction, and WithNotify pairs with a Listener for event-driven polling.
package postgres

import (
	"context"
	"fmt"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zalgonoise/cfg"

	"github.com/zalgonoise/x/gacha"
)

// DB is the subset of *pgxpool.Pool that the Repository uses.
//
// Use a *pgxpool.Pool for a repository that pollers claim from: claims must commit before the handlers run, and a
// poller calls the repository concurrently (heartbeats, per-entry settles). *pgx.Conn and pgx.Tx implement DB too,
// but only suit single-goroutine use; to write within a transaction, prefer Tx.
type DB interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Repository is a PostgreSQL outbox. It implements gacha.Repository, gacha.Extender, gacha.Redriver, gacha.Stater and
// gacha.Viewer. It is safe for concurrent use when its DB is, as a *pgxpool.Pool is.
type Repository struct {
	db      DB
	schema  string
	channel string
	q       queries
}

// New creates a Repository over db, for the outbox table named table. The name is required, so that streams of
// different item types never share a table by accident, and may be schema-qualified ("schema.table"). New does not
// touch the database: call Migrate, or apply Schema, to create the tables.
func New(db DB, table string, opts ...cfg.Option[Config]) (*Repository, error) {
	config := cfg.Set(defaultConfig(), opts...)

	out, dlq, err := tables(table, config.dlqTable)
	if err != nil {
		return nil, err
	}

	return &Repository{
		db:      db,
		schema:  replacer(out, dlq).Replace(schemaTemplate),
		channel: config.channel,
		q:       newQueries(out, dlq, config),
	}, nil
}

// Migrate creates the outbox table, the dead-letter table and their indexes if they do not exist. Concurrent calls
// are serialized with an advisory lock, so every instance of a service can call Migrate on startup.
func (r *Repository) Migrate(ctx context.Context) error {
	if _, err := r.db.Exec(ctx, migrateLock+r.schema); err != nil {
		return fmt.Errorf("%s: migrate: %w", domain, err)
	}

	return nil
}

// Write stores data as new pending entries in one statement, available for claims immediately. With WithNotify, it
// also notifies the channel.
func (r *Repository) Write(ctx context.Context, data ...[]byte) error {
	return r.write(ctx, r.db, data)
}

func (r *Repository) write(ctx context.Context, db DB, data [][]byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	ids := make([]pgtype.UUID, len(data))

	for i := range ids {
		ids[i] = pgtype.UUID{Bytes: uuid.NewV7(), Valid: true}
	}

	args := []any{ids, nonNil(data)}
	if r.channel != "" {
		args = append(args, r.channel)
	}

	if _, err := db.Exec(ctx, r.q.write, args...); err != nil {
		return fmt.Errorf("%s: write: %w", domain, err)
	}

	return nil
}

// Claim leases up to limit claimable entries for the lease duration, as described by gacha.Source.
func (r *Repository) Claim(ctx context.Context, limit int, lease time.Duration) ([]gacha.Entry[[]byte], error) {
	return r.claim(ctx, r.q.claim, "claim", limit, lease)
}

func (r *Repository) claim(
	ctx context.Context, query, op string, limit int, lease time.Duration, args ...any,
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

	rows, err := r.db.Query(ctx, query, append([]any{limit, leaseInterval(lease)}, args...)...)
	if err != nil {
		return nil, fmt.Errorf("%s: %s: %w", domain, op, err)
	}

	entries, err := pgx.CollectRows(rows, scanEntry)
	if err != nil {
		return nil, fmt.Errorf("%s: %s: %w", domain, op, err)
	}

	return entries, nil
}

func scanEntry(row pgx.CollectableRow) (gacha.Entry[[]byte], error) {
	var entry gacha.Entry[[]byte]

	// Scanning into *[16]byte takes pgx's direct path, unlike the named uuid.UUID type.
	err := row.Scan((*[16]byte)(&entry.ID), &entry.Data, &entry.Attempts, &entry.CreatedAt)

	return entry, err
}

// Settle applies the outcomes atomically in one statement, as described by gacha.Source. When several outcomes share
// an ID, the first one wins.
func (r *Repository) Settle(ctx context.Context, outcomes ...gacha.Outcome) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := validate(outcomes); err != nil {
		return err
	}

	if len(outcomes) == 0 {
		return nil
	}

	var args settleArgs

	for _, outcome := range dedupe(outcomes) {
		args.add(&outcome)
	}

	if _, err := r.db.Exec(ctx, r.q.settle, args.values()...); err != nil {
		return fmt.Errorf("%s: settle: %w", domain, err)
	}

	return nil
}

// Extend renews the leases of the given processing entries, as described by gacha.Extender.
func (r *Repository) Extend(ctx context.Context, lease time.Duration, ids ...uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if lease <= 0 {
		return gacha.ErrInvalidLease
	}

	if len(ids) == 0 {
		return nil
	}

	if _, err := r.db.Exec(ctx, r.q.extend, leaseInterval(lease), pgUUIDs(ids)); err != nil {
		return fmt.Errorf("%s: extend: %w", domain, err)
	}

	return nil
}

// Redrive moves the given dead letters back to the outbox in one statement, as described by gacha.Redriver. With
// WithNotify, it also notifies the channel.
func (r *Repository) Redrive(ctx context.Context, ids ...uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if len(ids) == 0 {
		return nil
	}

	args := []any{pgUUIDs(ids)}
	if r.channel != "" {
		args = append(args, r.channel)
	}

	if _, err := r.db.Exec(ctx, r.q.redrive, args...); err != nil {
		return fmt.Errorf("%s: redrive: %w", domain, err)
	}

	return nil
}

// Stats returns the number of entries in each state. It counts every row of both tables, so it suits monitoring
// at a modest rate rather than hot paths.
func (r *Repository) Stats(ctx context.Context) (gacha.Stats, error) {
	if err := ctx.Err(); err != nil {
		return gacha.Stats{}, err
	}

	var stats gacha.Stats

	err := r.db.QueryRow(ctx, r.q.stats).Scan(&stats.Pending, &stats.Processing, &stats.Completed, &stats.DeadLetters)
	if err != nil {
		return gacha.Stats{}, fmt.Errorf("%s: stats: %w", domain, err)
	}

	return stats, nil
}

// Tx returns a sink that writes to the outbox within tx, so the entries commit or roll back together with the
// caller's business data. With WithNotify, the notification is delivered when tx commits.
func (r *Repository) Tx(tx pgx.Tx) gacha.Sink[[]byte] {
	return gacha.SinkFunc[[]byte](func(ctx context.Context, data ...[]byte) error {
		return r.write(ctx, tx, data)
	})
}

// nonNil returns data with nil values replaced by empty ones, since the data column is NOT NULL. It only copies data
// when it holds nil values.
func nonNil(data [][]byte) [][]byte {
	for i := range data {
		if data[i] != nil {
			continue
		}

		out := make([][]byte, len(data))
		copy(out, data)

		for j := i; j < len(out); j++ {
			if out[j] == nil {
				out[j] = []byte{}
			}
		}

		return out
	}

	return data
}

func validate(outcomes []gacha.Outcome) error {
	for i := range outcomes {
		if err := outcomes[i].Validate(); err != nil {
			return err
		}
	}

	return nil
}

// dedupe returns the outcomes without repeated IDs, keeping the first outcome of each ID: a statement cannot modify
// the same row twice.
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

// pgUUIDs converts ids for pgx, whose []pgtype.UUID encoding avoids reflection.
func pgUUIDs(ids []uuid.UUID) []pgtype.UUID {
	out := make([]pgtype.UUID, len(ids))

	for i := range ids {
		out[i] = pgUUID(ids[i])
	}

	return out
}

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func pgInterval(d time.Duration) pgtype.Interval {
	return pgtype.Interval{Microseconds: max(d, 0).Microseconds(), Valid: true}
}

// leaseInterval rounds a positive lease up to whole microseconds, the database's resolution, so that a short lease
// never becomes an immediately expired one.
func leaseInterval(lease time.Duration) pgtype.Interval {
	return pgtype.Interval{Microseconds: int64((lease + time.Microsecond - 1) / time.Microsecond), Valid: true}
}

// settleArgs holds the array parameters of the settle statement, one group per outcome status. Their types are the
// ones pgx encodes without reflection.
type settleArgs struct {
	completedIDs      []pgtype.UUID
	completedAttempts []int32
	completedErrs     []string
	retriedIDs        []pgtype.UUID
	retriedAttempts   []int32
	retriedErrs       []string
	retriedIn         []pgtype.Interval
	failedIDs         []pgtype.UUID
	failedAttempts    []int32
	failedErrs        []string
	releasedIDs       []pgtype.UUID
	releasedAttempts  []int32
	releasedErrs      []string
}

func (a *settleArgs) add(outcome *gacha.Outcome) {
	attempts := int32(min(outcome.Attempts, maxInt32)) //nolint:gosec // Clamped to the int4 column's range.

	switch outcome.Status {
	case gacha.StatusCompleted:
		a.completedIDs = append(a.completedIDs, pgUUID(outcome.ID))
		a.completedAttempts = append(a.completedAttempts, attempts)
		a.completedErrs = append(a.completedErrs, outcome.Err)
	case gacha.StatusPending:
		a.retriedIDs = append(a.retriedIDs, pgUUID(outcome.ID))
		a.retriedAttempts = append(a.retriedAttempts, attempts)
		a.retriedErrs = append(a.retriedErrs, outcome.Err)
		a.retriedIn = append(a.retriedIn, pgInterval(outcome.RetryIn))
	case gacha.StatusFailed:
		a.failedIDs = append(a.failedIDs, pgUUID(outcome.ID))
		a.failedAttempts = append(a.failedAttempts, attempts)
		a.failedErrs = append(a.failedErrs, outcome.Err)
	case gacha.StatusReleased:
		a.releasedIDs = append(a.releasedIDs, pgUUID(outcome.ID))
		a.releasedAttempts = append(a.releasedAttempts, attempts)
		a.releasedErrs = append(a.releasedErrs, outcome.Err)
	case gacha.StatusProcessing:
		// Rejected by validate.
	}
}

func (a *settleArgs) values() []any {
	return []any{
		a.completedIDs, a.completedAttempts, a.completedErrs,
		a.retriedIDs, a.retriedAttempts, a.retriedErrs, a.retriedIn,
		a.failedIDs, a.failedAttempts, a.failedErrs,
		a.releasedIDs, a.releasedAttempts, a.releasedErrs,
	}
}

const maxInt32 = 1<<31 - 1
