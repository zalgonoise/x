package database

import (
	"context"
	"database/sql"
	"log/slog"
)

func OpenPostgres(ctx context.Context, uri string, maxConns int, logger *slog.Logger) (*sql.DB, error) {
	db, err := sql.Open("pgx", uri)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxConns)
	db.SetMaxIdleConns(maxConns)

	if err = Migrate(ctx, db, logger); err != nil {
		return nil, err
	}

	return db, nil
}
