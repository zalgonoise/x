package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

const minBlockSize = 32_000

type migration struct {
	table  string
	create string
}

func MigrateSQLite(ctx context.Context, db *sql.DB, logger *slog.Logger) error {
	start := time.Now()

	if err := runMigrations(ctx, db, newSchema()...); err != nil {
		return err
	}

	logger.InfoContext(ctx, "operation completed", slog.Duration("time_elapsed", time.Since(start)))

	return nil
}

func runMigrations(ctx context.Context, db *sql.DB, migrations ...migration) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	for i := range migrations {
		r, err := tx.QueryContext(ctx, fmt.Sprintf(checkTableExists, migrations[i].table))
		if err != nil {
			return err
		}

		var count int

		if !r.Next() {
			return r.Err()
		}

		if err = r.Scan(&count); err != nil {
			_ = r.Close()

			return err
		}

		_ = r.Close()

		if count == 1 {
			continue
		}

		_, err = tx.ExecContext(ctx, migrations[i].create)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
