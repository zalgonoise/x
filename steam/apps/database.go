package apps

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/zalgonoise/x/steam"
)

const (
	uriFormat = "file:%s?cache=shared"
	inMemory  = ":memory:"
)

func open(config Config) (*sql.DB, error) {
	if config.uri != inMemory {
		if err := validateURI(config.uri); err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("sqlite", fmt.Sprintf(uriFormat, config.uri))
	if err != nil {
		return nil, err
	}

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

func initDatabase(db *sql.DB) error {
	ctx := context.Background()
	r, err := db.QueryContext(ctx, checkTableExists)
	if err != nil {
		return err
	}

	defer r.Close()

	for r.Next() {
		var count int
		if err = r.Scan(&count); err != nil {
			return err
		}

		if count == 1 {
			return nil
		}
	}

	_, err = db.ExecContext(ctx, createTableQuery)
	if err != nil {
		return err
	}

	apps, err := steam.LoadAppsList()
	if err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	for i := range apps {
		if _, err = tx.ExecContext(ctx, insertValueQuery, apps[i].AppID, apps[i].Name); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return tx.Rollback()
	}

	return nil
}
