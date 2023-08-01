package sqlite

import (
	"context"
	"database/sql"
	"time"

	"go.opentelemetry.io/otel/trace"
)

const (
	minAlloc = 16
)

const (
	insertTraceQuery = `
INSERT INTO traces (
    traceID, timestamp
) VALUES (?, ?)
`

	listTracesQuery = `
SELECT traceID FROM traces
WHERE timestamp < ?
`

	deleteTracesQuery = `
DELETE FROM traces
WHERE timestamp < ?
`
)

type dbTrace struct {
	traceID   sql.NullString
	timestamp sql.NullInt64
}

func toDBTrace(traceID trace.TraceID) dbTrace {
	return dbTrace{
		traceID:   sql.NullString{String: string(traceID[:]), Valid: true},
		timestamp: sql.NullInt64{Int64: time.Now().Unix(), Valid: true},
	}
}

// Repository implements logbuf.Repository
type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return Repository{db}
}

// InsertTrace adds the input trace.TraceID to the database if it does not yet exist, alongside with the current
// timestamp (of when it is registered). Returns an error if raised.
func (r Repository) InsertTrace(ctx context.Context, traceID trace.TraceID) (err error) {
	t := toDBTrace(traceID)

	if _, err = r.db.ExecContext(ctx, insertTraceQuery, t.traceID, t.timestamp); err != nil {
		return err
	}

	return nil
}

// DeleteTraces removes all trace.TraceID from the database that are older than the threshold time.Duration
// (which is calculated from the current time minus this value). It returns a slice of trace.TraceID with all
// values that are were removed and an error if raised.
func (r Repository) DeleteTraces(ctx context.Context, threshold time.Duration) (pruned []trace.TraceID, err error) {
	limit := time.Now().Add(-threshold)

	unixTime := sql.NullInt64{Int64: limit.Unix(), Valid: !limit.IsZero()}

	rows, err := r.db.QueryContext(ctx, listTracesQuery, unixTime)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	pruned = make([]trace.TraceID, 0, minAlloc)

	for rows.Next() {
		var id string

		if err = rows.Scan(&id); err != nil {
			return nil, err
		}

		pruned = append(pruned, [16]byte([]byte(id)))
	}

	if err = rows.Close(); err != nil {
		return nil, err
	}

	// Rows.Err will report the last error encountered by Rows.Scan.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	if _, err = r.db.ExecContext(ctx, deleteTracesQuery, unixTime); err != nil {
		return nil, err
	}

	return pruned, nil
}
