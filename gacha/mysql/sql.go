package mysql

import (
	"strings"
	"time"
	"uuid"
)

// now is the database time used by every statement. UTC_TIMESTAMP does not depend on the session's time zone, and
// MySQL evaluates it once per statement, so all the rows of a statement share it.
const now = "UTC_TIMESTAMP(6)"

const (
	// maxRows caps the rows (or outcomes, or IDs) per statement, well below the 65535 placeholders limit.
	maxRows = 1000
	// maxBytes caps the data written per INSERT statement, well below max_allowed_packet's defaults (16MiB on
	// MariaDB, 64MiB on MySQL).
	maxBytes = 4 << 20
)

// The rows of the derived tables of parameters used by settles. Casting every parameter types the derived tables the
// same way on MySQL and MariaDB, with or without interpolateParams.
const (
	outcomeRow = "SELECT CAST(? AS BINARY(16)) AS id, CAST(? AS SIGNED) AS attempts, CAST(? AS CHAR) AS err"
	retryRow   = "SELECT CAST(? AS BINARY(16)) AS id, CAST(? AS SIGNED) AS attempts, CAST(? AS CHAR) AS err, " +
		"CAST(? AS SIGNED) AS retry_us"
	delayRow = "SELECT CAST(? AS BINARY(16)) AS id, CAST(? AS SIGNED) AS retry_us"

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

// values returns a derived table of n rows, built with UNION ALL, which both MySQL and MariaDB support.
func values(row string, n int) string {
	return "(" + row + strings.Repeat(" UNION ALL "+row, n-1) + ")"
}

// interval returns d in microseconds, for INTERVAL ? MICROSECOND, treating negative durations as zero.
func interval(d time.Duration) int64 {
	return max(d, 0).Microseconds()
}

// leaseMicros rounds a positive lease up to whole microseconds, the database's resolution, so that a short lease
// never becomes an immediately expired one.
func leaseMicros(lease time.Duration) int64 {
	return int64((lease + time.Microsecond - 1) / time.Microsecond)
}

// idArgs appends the IDs to args as byte slices, for BINARY(16) columns.
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
