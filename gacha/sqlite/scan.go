package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/zalgonoise/x/gacha"
)

// idScanner scans a BLOB id column into a UUID.
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

// timeScanner scans an INTEGER column of microseconds since the Unix epoch.
type timeScanner struct {
	t *time.Time
}

func (s timeScanner) Scan(src any) error {
	micros, ok := src.(int64)
	if !ok {
		return fmt.Errorf("%w: time %v", ErrInvalidValue, src)
	}

	*s.t = time.UnixMicro(micros).UTC()

	return nil
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

		if entry.Data == nil {
			entry.Data = []byte{}
		}

		entries = append(entries, entry)
	}

	return entries, rows.Err()
}
