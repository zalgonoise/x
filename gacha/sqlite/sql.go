package sqlite

import (
	"strings"
	"time"
	"uuid"
)

// now is the database time used by every statement, in microseconds since the Unix epoch. julianday('now') has
// millisecond precision, so the value is rounded to the millisecond. SQLite evaluates 'now' once per statement, so
// all the rows of a statement share it.
const now = "(CAST(ROUND((julianday('now') - 2440587.5) * 86400000) AS INTEGER) * 1000)"

// maxRows caps the rows (or outcomes, or IDs) per statement, well below SQLite's default limit of 32766 parameters.
const maxRows = 1000

// The rows of the CTEs of parameters used by settles.
const (
	outcomeColumns = "i(id, attempts, err)"
	outcomeRow     = "(?, ?, ?)"
	retryColumns   = "i(id, attempts, err, retry_us)"
	retryRow       = "(?, ?, ?, ?)"
	delayColumns   = "i(id, retry_us)"
	delayRow       = "(?, ?)"

	outcomeArgs = 3
	retryArgs   = 4
	delayArgs   = 2
	insertArgs  = 2
)

// statement is a query and its arguments.
type statement struct {
	query string
	args  []any
}

// placeholders returns n comma-separated placeholders.
func placeholders(n int) string {
	return strings.TrimSuffix(strings.Repeat("?, ", n), ", ")
}

// with returns a WITH clause naming n rows of parameters, such as WITH i(id, retry_us) AS (VALUES (?, ?), (?, ?)).
func with(columns, row string, n int) string {
	return "WITH " + columns + " AS (VALUES " + row + strings.Repeat(", "+row, n-1) + ") "
}

// micros returns d in microseconds, treating negative durations as zero.
func micros(d time.Duration) int64 {
	return max(d, 0).Microseconds()
}

// leaseMicros rounds a positive lease up to whole microseconds, so that a short lease never becomes an immediately
// expired one. Leases are compared to the millisecond clock: a sub-millisecond lease expires within the next
// millisecond.
func leaseMicros(lease time.Duration) int64 {
	return int64((lease + time.Microsecond - 1) / time.Microsecond)
}

// idArgs appends the IDs to args as byte slices, for the BLOB id columns.
func idArgs(args []any, ids []uuid.UUID) []any {
	for i := range ids {
		args = append(args, ids[i][:])
	}

	return args
}

// spans splits n items into consecutive [from, to) ranges of at most maxRows items, one per statement.
func spans(n int) [][2]int {
	ranges := make([][2]int, 0, (n+maxRows-1)/maxRows)

	for from := 0; from < n; from += maxRows {
		ranges = append(ranges, [2]int{from, min(from+maxRows, n)})
	}

	return ranges
}
