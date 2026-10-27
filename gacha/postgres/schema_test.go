package postgres

import (
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/gacha"
)

// Compile-time checks that Repository implements every gacha capability.
var (
	_ gacha.Repository[[]byte] = (*Repository)(nil)
	_ gacha.Extender           = (*Repository)(nil)
	_ gacha.Redriver           = (*Repository)(nil)
	_ gacha.Stater             = (*Repository)(nil)
	_ gacha.Viewer[[]byte]     = (*Repository)(nil)
)

func TestSchema(t *testing.T) {
	for _, testcase := range []struct {
		name       string
		table      string
		dlqTable   string
		wantsTable string
		wantsDLQ   string
		wantsIndex string
		err        error
	}{
		{
			name:       "Simple",
			table:      "orders_outbox",
			wantsTable: `CREATE TABLE IF NOT EXISTS "orders_outbox" (`,
			wantsDLQ:   `CREATE TABLE IF NOT EXISTS "orders_outbox_dlq" (`,
			wantsIndex: `CREATE INDEX IF NOT EXISTS "orders_outbox_claim_idx"`,
		},
		{
			name:       "SchemaQualified",
			table:      "billing.orders_outbox",
			wantsTable: `CREATE TABLE IF NOT EXISTS "billing"."orders_outbox" (`,
			wantsDLQ:   `CREATE TABLE IF NOT EXISTS "billing"."orders_outbox_dlq" (`,
			wantsIndex: `CREATE INDEX IF NOT EXISTS "orders_outbox_claim_idx"`,
		},
		{
			name:       "CustomDeadLetterTable",
			table:      "orders_outbox",
			dlqTable:   "archive.dead_orders",
			wantsTable: `CREATE TABLE IF NOT EXISTS "orders_outbox" (`,
			wantsDLQ:   `CREATE TABLE IF NOT EXISTS "archive"."dead_orders" (`,
			wantsIndex: `CREATE INDEX IF NOT EXISTS "dead_orders_failed_idx"`,
		},
		{
			name:       "QuotesIdentifiers",
			table:      `Weird"Name`,
			wantsTable: `CREATE TABLE IF NOT EXISTS "Weird""Name" (`,
			wantsDLQ:   `CREATE TABLE IF NOT EXISTS "Weird""Name_dlq" (`,
			wantsIndex: `CREATE INDEX IF NOT EXISTS "Weird""Name_completed_idx"`,
		},
		{name: "Empty", table: "", err: ErrInvalidTable},
		{name: "EmptyPart", table: "billing.", err: ErrInvalidTable},
		{name: "TooManyParts", table: "db.billing.orders", err: ErrInvalidTable},
		{name: "InvalidDeadLetterTable", table: "orders", dlqTable: "a.b.c", err: ErrInvalidTable},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			schema, err := Schema(testcase.table, testcase.dlqTable)
			if testcase.err != nil {
				require.ErrorIs(t, err, testcase.err)

				_, err = New(nil, testcase.table, WithDeadLetterTable(testcase.dlqTable))
				require.ErrorIs(t, err, testcase.err)

				return
			}

			require.NoError(t, err)
			require.Contains(t, schema, testcase.wantsTable)
			require.Contains(t, schema, testcase.wantsDLQ)
			require.Contains(t, schema, testcase.wantsIndex)
			require.NotContains(t, schema, "{", "every placeholder is replaced")

			repo, err := New(nil, testcase.table, WithDeadLetterTable(testcase.dlqTable))
			require.NoError(t, err)
			require.Equal(t, schema, repo.schema)
		})
	}
}

func TestQueries(t *testing.T) {
	out, dlq, err := tables("orders", "")
	require.NoError(t, err)

	for _, testcase := range []struct {
		name          string
		config        Config
		wantsNotify   bool
		wantsDeletion bool
	}{
		{name: "Default"},
		{name: "Notify", config: Config{channel: "orders"}, wantsNotify: true},
		{name: "DeleteOnComplete", config: Config{deleteOnComplete: true}, wantsDeletion: true},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			q := newQueries(out, dlq, testcase.config)

			require.Equal(t, testcase.wantsNotify, strings.Contains(q.write, "pg_notify"))
			require.Equal(t, testcase.wantsNotify, strings.Contains(q.redrive, "pg_notify"))
			require.Equal(t, testcase.wantsDeletion, strings.Contains(q.settle, "completed AS (\n\tDELETE"))

			for _, query := range []string{
				q.write, q.claim, q.settle, q.extend, q.redrive, q.stats,
				q.completedClaim, q.completedSettle, q.deadLettersClaim, q.deadLettersSettle,
			} {
				require.NotContains(t, query, "{")
			}
		})
	}
}

func TestDedupe(t *testing.T) {
	a, b := uuid.NewV7(), uuid.NewV7()

	for _, testcase := range []struct {
		name     string
		outcomes []gacha.Outcome
		wants    []gacha.Outcome
	}{
		{name: "Empty"},
		{
			name:     "Single",
			outcomes: []gacha.Outcome{{ID: a, Status: gacha.StatusCompleted}},
			wants:    []gacha.Outcome{{ID: a, Status: gacha.StatusCompleted}},
		},
		{
			name: "FirstWins",
			outcomes: []gacha.Outcome{
				{ID: a, Status: gacha.StatusCompleted},
				{ID: b, Status: gacha.StatusPending},
				{ID: a, Status: gacha.StatusFailed},
			},
			wants: []gacha.Outcome{
				{ID: a, Status: gacha.StatusCompleted},
				{ID: b, Status: gacha.StatusPending},
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			input := append([]gacha.Outcome(nil), testcase.outcomes...)

			require.Equal(t, testcase.wants, dedupe(input))
			require.Equal(t, testcase.outcomes, input, "the input is not modified")
		})
	}
}

func TestNonNil(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		data  [][]byte
		wants [][]byte
		same  bool
	}{
		{name: "NoNil", data: [][]byte{[]byte("a"), {}}, wants: [][]byte{[]byte("a"), {}}, same: true},
		{name: "Nil", data: [][]byte{[]byte("a"), nil, nil}, wants: [][]byte{[]byte("a"), {}, {}}},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			out := nonNil(testcase.data)

			for i := range out {
				require.NotNil(t, out[i])
				require.Equal(t, testcase.wants[i], out[i])
			}

			require.Equal(t, testcase.same, &out[0] == &testcase.data[0], "copies only when needed")
		})
	}
}

func TestSettleArgs(t *testing.T) {
	ids := []uuid.UUID{uuid.NewV7(), uuid.NewV7(), uuid.NewV7(), uuid.NewV7()}

	var args settleArgs

	for _, outcome := range []gacha.Outcome{
		{ID: ids[0], Attempts: 1, Status: gacha.StatusCompleted},
		{ID: ids[1], Attempts: 2, Status: gacha.StatusPending, Err: "retry", RetryIn: -time.Second},
		{ID: ids[2], Attempts: 3, Status: gacha.StatusFailed, Err: "boom"},
		{ID: ids[3], Attempts: 4, Status: gacha.StatusReleased},
	} {
		args.add(&outcome)
	}

	require.Equal(t, []any{
		pgUUIDs(ids[:1]), []int32{1}, []string{""},
		pgUUIDs(ids[1:2]), []int32{2}, []string{"retry"}, []pgtype.Interval{{Microseconds: 0, Valid: true}},
		pgUUIDs(ids[2:3]), []int32{3}, []string{"boom"},
		pgUUIDs(ids[3:]), []int32{4}, []string{""},
	}, args.values())
}

func TestNewListener(t *testing.T) {
	_, err := NewListener(nil, "")
	require.ErrorIs(t, err, ErrInvalidChannel)

	listener, err := NewListener(nil, "orders", WithReconnect(nil), WithLogger(nil))
	require.NoError(t, err)
	require.NotNil(t, listener.config.backoff)
	require.NotNil(t, listener.config.logger)
}
