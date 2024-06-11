package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"

	_ "modernc.org/sqlite"
)

const (
	sqlDriver = "sqlite"
	maxAlloc  = 5_000_000

	uriFormat = "file:%s?_readonly=true&_txlock=immediate&cache=shared"
	inMemory  = ":memory:"

	applyPragma   = `PRAGMA %s;`
	applyPragmaKV = `PRAGMA %s = %s;`

	checkTableExists = `
SELECT EXISTS(SELECT 1 FROM sqlite_master 
	WHERE type='table' 
	AND name='%s');
`

	insertScopesQuery = `
INSERT INTO scopes (id, min, max, total)
VALUES (?, ?, ?, ?);
`
)

func OpenSQLite(uri string, pragmas map[string]string, logger *slog.Logger) (*sql.DB, error) {
	switch uri {
	case inMemory:
	case "":
		uri = inMemory
	default:
		if err := validateURI(uri); err != nil {
			return nil, err
		}
	}

	if pragmas == nil {
		pragmas = ReadWritePragmas()
	}

	db, err := sql.Open(sqlDriver, fmt.Sprintf(uriFormat, uri))
	if err != nil {
		return nil, err
	}

	logger.Info("opened target DB", slog.String("uri", uri))

	if err := applyPragmas(context.Background(), db, pragmas); err != nil {
		return nil, err
	}

	logger.Info("prepared pragmas")

	return db, nil
}

func validateURI(uri string) error {
	stat, err := os.Stat(uri)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			f, err := os.Create(uri)
			if err != nil {
				return err
			}

			return f.Close()
		}

		return err
	}

	if stat.IsDir() {
		return fmt.Errorf("%s is a directory", uri)
	}

	return nil
}

func applyPragmas(ctx context.Context, db *sql.DB, pragmas map[string]string) (err error) {
	for k, v := range pragmas {
		switch v {
		case "":
			_, err = db.ExecContext(ctx, fmt.Sprintf(applyPragma, k))
		default:
			_, err = db.ExecContext(ctx, fmt.Sprintf(applyPragmaKV, k, v))
		}

		if err != nil {
			return err
		}
	}

	return nil
}
