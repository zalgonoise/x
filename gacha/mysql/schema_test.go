package mysql

import (
	"strings"
	"testing"
	"time"
	"uuid"

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
		err        error
	}{
		{
			name:       "Simple",
			table:      "orders_outbox",
			wantsTable: "CREATE TABLE IF NOT EXISTS `orders_outbox` (",
			wantsDLQ:   "CREATE TABLE IF NOT EXISTS `orders_outbox_dlq` (",
		},
		{
			name:       "DatabaseQualified",
			table:      "billing.orders_outbox",
			wantsTable: "CREATE TABLE IF NOT EXISTS `billing`.`orders_outbox` (",
			wantsDLQ:   "CREATE TABLE IF NOT EXISTS `billing`.`orders_outbox_dlq` (",
		},
		{
			name:       "CustomDeadLetterTable",
			table:      "orders_outbox",
			dlqTable:   "archive.dead_orders",
			wantsTable: "CREATE TABLE IF NOT EXISTS `orders_outbox` (",
			wantsDLQ:   "CREATE TABLE IF NOT EXISTS `archive`.`dead_orders` (",
		},
		{
			name:       "QuotesIdentifiers",
			table:      "weird`name",
			wantsTable: "CREATE TABLE IF NOT EXISTS `weird``name` (",
			wantsDLQ:   "CREATE TABLE IF NOT EXISTS `weird``name_dlq` (",
		},
		{name: "Empty", table: "", err: ErrInvalidTable},
		{name: "EmptyPart", table: ".orders", err: ErrInvalidTable},
		{name: "TooManyParts", table: "a.b.c", err: ErrInvalidTable},
		{name: "InvalidDeadLetterTable", table: "orders", dlqTable: "a..b", err: ErrInvalidTable},
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
			require.NotContains(t, schema, "{")
			require.Equal(t, 2, strings.Count(schema, ";"))

			repo, err := New(nil, testcase.table, WithDeadLetterTable(testcase.dlqTable))
			require.NoError(t, err)
			require.Equal(t, schema, strings.Join(repo.schema, ";\n\n")+";\n")
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
		{name: "Values/One", got: values(delayRow, 1), wants: "(" + delayRow + ")"},
		{name: "Values/Two", got: values(delayRow, 2), wants: "(" + delayRow + " UNION ALL " + delayRow + ")"},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			require.Equal(t, testcase.wants, testcase.got)
		})
	}

	require.Equal(t, [][2]int{}, spans(0))
	require.Equal(t, [][2]int{{0, 5}}, spans(5))
	require.Equal(t, [][2]int{{0, maxRows}, {maxRows, 2 * maxRows}, {2 * maxRows, 2*maxRows + 1}}, spans(2*maxRows+1))
	require.Equal(t, int64(1500), interval(1500*time.Microsecond))
	require.Equal(t, int64(0), interval(-time.Second))
}

func TestInserts(t *testing.T) {
	repo, err := New(nil, "orders")
	require.NoError(t, err)

	large := make([]byte, maxBytes/2+1)

	for _, testcase := range []struct {
		name       string
		data       [][]byte
		wantsSizes []int
	}{
		{name: "Empty"},
		{name: "Single", data: [][]byte{[]byte("a")}, wantsSizes: []int{1}},
		{name: "ByRows", data: make([][]byte, 2*maxRows+1), wantsSizes: []int{maxRows, maxRows, 1}},
		{name: "ByBytes", data: [][]byte{large, large, large}, wantsSizes: []int{1, 1, 1}},
		{name: "OversizedRowAlone", data: [][]byte{[]byte("a"), make([]byte, maxBytes+1), []byte("b")}, wantsSizes: []int{1, 1, 1}},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			statements := repo.inserts(testcase.data)
			require.Len(t, statements, len(testcase.wantsSizes))

			for i, size := range testcase.wantsSizes {
				require.Len(t, statements[i].args, insertArgs*size)
				require.Equal(t, size, strings.Count(statements[i].query, "(?, ?, "))

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
	wantsTime := time.Date(2026, 10, 23, 11, 30, 0, 123456000, time.UTC)

	for _, testcase := range []struct {
		name    string
		scanner interface{ Scan(src any) error }
		src     any
		err     error
	}{
		{name: "ID", scanner: idScanner{new(uuid.UUID)}, src: id[:]},
		{name: "ID/WrongLength", scanner: idScanner{new(uuid.UUID)}, src: []byte("short"), err: ErrInvalidValue},
		{name: "ID/WrongType", scanner: idScanner{new(uuid.UUID)}, src: int64(1), err: ErrInvalidValue},
		{name: "Time/Time", scanner: timeScanner{new(time.Time)}, src: wantsTime},
		{
			name:    "Time/TimeInOtherLocation",
			scanner: timeScanner{new(time.Time)},
			src:     time.Date(2026, 10, 23, 11, 30, 0, 123456000, time.FixedZone("UTC+2", 2*60*60)),
		},
		{name: "Time/Bytes", scanner: timeScanner{new(time.Time)}, src: []byte("2026-10-23 11:30:00.123456")},
		{name: "Time/String", scanner: timeScanner{new(time.Time)}, src: "2026-10-23 11:30:00.123456"},
		{name: "Time/WrongType", scanner: timeScanner{new(time.Time)}, src: int64(1), err: ErrInvalidValue},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			err := testcase.scanner.Scan(testcase.src)
			if testcase.err != nil {
				require.ErrorIs(t, err, testcase.err)

				return
			}

			require.NoError(t, err)

			switch s := testcase.scanner.(type) {
			case idScanner:
				require.Equal(t, id, *s.id)
			case timeScanner:
				require.True(t, wantsTime.Equal(*s.t), "got %s", *s.t)
			}
		})
	}
}
