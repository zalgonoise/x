package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/zalgonoise/x/gacha"
)

// datetimeLayout is the text form of DATETIME(6) values, returned when the DSN does not set parseTime.
const datetimeLayout = "2006-01-02 15:04:05.999999"

// idScanner scans a BINARY(16) column into a UUID.
type idScanner struct {
	id *uuid.UUID
}

func (s idScanner) Scan(src any) error {
	b, ok := src.([]byte)
	if !ok || len(b) != len(s.id) {
		return fmt.Errorf("%w: id %v", ErrInvalidValue, src)
	}

	copy(s.id[:], b)

	return nil
}

// timeScanner scans a DATETIME(6) column holding UTC times, whether or not the DSN sets parseTime.
type timeScanner struct {
	t *time.Time
}

func (s timeScanner) Scan(src any) error {
	var err error

	switch v := src.(type) {
	case time.Time:
		// With parseTime, the driver reads the wall clock in the DSN's loc: it is UTC, whatever loc says.
		*s.t = time.Date(v.Year(), v.Month(), v.Day(), v.Hour(), v.Minute(), v.Second(), v.Nanosecond(), time.UTC)
	case []byte:
		*s.t, err = time.ParseInLocation(datetimeLayout, string(v), time.UTC)
	case string:
		*s.t, err = time.ParseInLocation(datetimeLayout, v, time.UTC)
	default:
		err = fmt.Errorf("%w: time %v", ErrInvalidValue, src)
	}

	return err
}

// scanEntries reads entries from rows of (id, data, attempts, created_at), closing rows.
func scanEntries(rows *sql.Rows, capacity int) (entries []gacha.Entry[[]byte], err error) {
	defer func() {
		err = errors.Join(err, rows.Close())
	}()

	entries = make([]gacha.Entry[[]byte], 0, capacity)

	for rows.Next() {
		var entry gacha.Entry[[]byte]

		if err = rows.Scan(idScanner{&entry.ID}, &entry.Data, &entry.Attempts, timeScanner{&entry.CreatedAt}); err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	return entries, rows.Err()
}
