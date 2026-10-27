package sqlite

import (
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/gacha"
)

func TestSchema(t *testing.T) {
	for _, testcase := range []struct {
		name       string
		table      string
		dlqTable   string
		wantsTable string
		wantsIndex string
		wantsDLQ   string
		err        error
	}{
		{
			name:       "Simple",
			table:      "orders_outbox",
			wantsTable: `CREATE TABLE IF NOT EXISTS "orders_outbox" (`,
			wantsIndex: `CREATE INDEX IF NOT EXISTS "orders_outbox_claim_idx"` + "\n" + `    ON "orders_outbox" (`,
			wantsDLQ:   `CREATE TABLE IF NOT EXISTS "orders_outbox_dlq" (`,
		},
		{
			name:       "AttachedDatabase",
			table:      "aux.orders_outbox",
			wantsTable: `CREATE TABLE IF NOT EXISTS "aux"."orders_outbox" (`,
			wantsIndex: `CREATE INDEX IF NOT EXISTS "aux"."orders_outbox_claim_idx"` + "\n" + `    ON "orders_outbox" (`,
			wantsDLQ:   `CREATE TABLE IF NOT EXISTS "aux"."orders_outbox_dlq" (`,
		},
		{
			name:       "CustomDeadLetterTable",
			table:      "orders_outbox",
			dlqTable:   "dead",
			wantsTable: `CREATE TABLE IF NOT EXISTS "orders_outbox" (`,
			wantsIndex: `CREATE INDEX IF NOT EXISTS "dead_failed_idx"` + "\n" + `    ON "dead" (`,
			wantsDLQ:   `CREATE TABLE IF NOT EXISTS "dead" (`,
		},
		{
			name:       "QuotesIdentifiers",
			table:      `weird"name`,
			wantsTable: `CREATE TABLE IF NOT EXISTS "weird""name" (`,
			wantsIndex: `CREATE INDEX IF NOT EXISTS "weird""name_completed_idx"`,
			wantsDLQ:   `CREATE TABLE IF NOT EXISTS "weird""name_dlq" (`,
		},
		{name: "Empty", table: "", err: ErrInvalidTable},
		{name: "EmptyPart", table: "aux.", err: ErrInvalidTable},
		{name: "TooManyParts", table: "a.b.c", err: ErrInvalidTable},
		{name: "InvalidDeadLetterTable", table: "orders", dlqTable: ".dead", err: ErrInvalidTable},
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
			require.Contains(t, schema, testcase.wantsIndex)
			require.Contains(t, schema, testcase.wantsDLQ)
			require.NotContains(t, schema, "{")
			require.NotContains(t, schema, ";;")
			require.Equal(t, 5, strings.Count(schema, ";\n"))

			repo, err := New(nil, testcase.table, WithDeadLetterTable(testcase.dlqTable))
			require.NoError(t, err)
			require.Len(t, repo.schema, 5)
		})
	}
}

func TestStatements(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		got   string
		wants string
	}{
		{name: "Placeholders/One", got: placeholders(1), wants: "?"},
		{name: "Placeholders/Three", got: placeholders(3), wants: "?, ?, ?"},
		{name: "With/One", got: with(delayColumns, delayRow, 1), wants: "WITH i(id, retry_us) AS (VALUES (?, ?)) "},
		{
			name:  "With/Two",
			got:   with(delayColumns, delayRow, 2),
			wants: "WITH i(id, retry_us) AS (VALUES (?, ?), (?, ?)) ",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			require.Equal(t, testcase.wants, testcase.got)
		})
	}

	require.Equal(t, [][2]int{}, spans(0))
	require.Equal(t, [][2]int{{0, maxRows}, {maxRows, maxRows + 1}}, spans(maxRows+1))
	require.Equal(t, int64(1500), micros(1500*time.Microsecond))
	require.Equal(t, int64(0), micros(-time.Second))
}

func TestInserts(t *testing.T) {
	repo, err := New(nil, "orders")
	require.NoError(t, err)

	for _, testcase := range []struct {
		name       string
		data       [][]byte
		wantsSizes []int
	}{
		{name: "Empty"},
		{name: "Single", data: [][]byte{[]byte("a")}, wantsSizes: []int{1}},
		{name: "Chunked", data: make([][]byte, 2*maxRows+1), wantsSizes: []int{maxRows, maxRows, 1}},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			statements := repo.inserts(testcase.data)
			require.Len(t, statements, len(testcase.wantsSizes))

			for i, size := range testcase.wantsSizes {
				require.Len(t, statements[i].args, insertArgs*size)
				require.Equal(t, size, strings.Count(statements[i].query, "(?, ?)"))

				for j := 1; j < len(statements[i].args); j += insertArgs {
					require.NotNil(t, statements[i].args[j], "nil data is written as empty")
				}
			}
		})
	}
}

func TestDedupeAndGroup(t *testing.T) {
	a, b, c, d := uuid.NewV7(), uuid.NewV7(), uuid.NewV7(), uuid.NewV7()

	completed, retried, failed, released := group(dedupe([]gacha.Outcome{
		{ID: a, Status: gacha.StatusCompleted},
		{ID: b, Status: gacha.StatusPending},
		{ID: a, Status: gacha.StatusFailed},
		{ID: c, Status: gacha.StatusFailed},
		{ID: d, Status: gacha.StatusReleased},
	}))

	require.Equal(t, []gacha.Outcome{{ID: a, Status: gacha.StatusCompleted}}, completed)
	require.Equal(t, []gacha.Outcome{{ID: b, Status: gacha.StatusPending}}, retried)
	require.Equal(t, []gacha.Outcome{{ID: c, Status: gacha.StatusFailed}}, failed)
	require.Equal(t, []gacha.Outcome{{ID: d, Status: gacha.StatusReleased}}, released)
}

func TestScanners(t *testing.T) {
	id := uuid.NewV7()
	at := time.Date(2026, 10, 23, 11, 30, 0, 123000000, time.UTC)

	var gotID uuid.UUID

	require.NoError(t, idScanner{&gotID}.Scan(id[:]))
	require.Equal(t, id, gotID)
	require.ErrorIs(t, idScanner{&gotID}.Scan([]byte("short")), ErrInvalidValue)
	require.ErrorIs(t, idScanner{&gotID}.Scan("text"), ErrInvalidValue)

	var gotTime time.Time

	require.NoError(t, timeScanner{&gotTime}.Scan(at.UnixMicro()))
	require.Equal(t, at, gotTime)
	require.ErrorIs(t, timeScanner{&gotTime}.Scan("2026-10-23"), ErrInvalidValue)
}
