package database

import (
	"context"
	"database/sql"
	"fmt"
)

const (
	checkTableExists = `
SELECT EXISTS(SELECT 1 FROM sqlite_master 
	WHERE type='table' 
	AND name='%s');
`
)

type migration struct {
	table  string
	create string
}

func MigrateSQLite(ctx context.Context, db *sql.DB) error {
	return runMigrations(ctx, db,
		migration{table: "headers", create: createHeadersTable},
		migration{table: "chunks", create: createChunksTable},
	)
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
