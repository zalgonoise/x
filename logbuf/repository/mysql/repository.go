package mysql

import (
	"context"
	"database/sql"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// Repository implements logbuf.Repository
type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return Repository{db}
}

// InsertTrace adds the input trace.TraceID to the database if it does not yet exist, alongside with the current
// timestamp (of when it is registered). Returns an error if raised.
//
// TODO: implement method
func (r Repository) InsertTrace(ctx context.Context, traceID trace.TraceID) (err error) {

	return nil
}

// DeleteTraces removes all trace.TraceID from the database that are older than the threshold time.Duration
// (which is calculated from the current time minus this value). It returns a slice of trace.TraceID with all
// values that are were removed and an error if raised.
//
// TODO: implement method
func (r Repository) DeleteTraces(ctx context.Context, threshold time.Duration) (pruned []trace.TraceID, err error) {

	return nil, nil
}
