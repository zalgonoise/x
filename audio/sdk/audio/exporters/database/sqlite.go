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
	sqlDriver = "sqlite3"

	uriFormat = "file:%s?_readonly=true&_txlock=immediate&cache=shared"
	inMemory  = ":memory:"

	applyPragma   = `PRAGMA %s;`
	applyPragmaKV = `PRAGMA %s = %s;`

	createHeadersTable = `
CREATE TABLE headers (
    uuid TEXT PRIMARY KEY NOT NULL,
    timestamp INTEGER NOT NULL,
    
    chunk_id BLOB NOT NULL,
    chunk_size INTEGER NOT NULL,
    format BLOB NOT NULL,
    subchunk_1_id BLOB NOT NULL,
    subchunk_1_size INTEGER NOT NULL,
    audio_format INTEGER NOT NULL,
    num_channels INTEGER NOT NULL,
    sample_rate INTEGER NOT NULL,
    byte_rate INTEGER NOT NULL,
    block_align INTEGER NOT NULL,
    bits_per_sample INTEGER NOT NULL
) STRICT;
`

	createChunksTable = `
CREATE TABLE chunks (
    uuid TEXT PRIMARY KEY NOT NULL,
    header_id TEXT REFERENCES headers (uuid) NOT NULL,
    timestamp INTEGER NOT NULL,
    
    subchunk_data BLOB
) STRICT;
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
