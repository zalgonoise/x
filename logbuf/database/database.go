package database

import (
	"context"
	"database/sql"
	"embed"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite"
)

//go:embed migration/*
var migrations embed.FS

type Logger interface {
	InfoContext(ctx context.Context, msg string, args ...any)
}

// Connect creates a new instance of a SQLite database from the input URI. This URI can be empty (or ":memory:", if the
// caller wishes to run the database in-memory.
func Connect(uri string) (*sql.DB, error) {
	if uri == "" {
		uri = ":memory:"
	}

	db, err := sql.Open("sqlite", uri)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Migrate updates an instance of a database to the latest schema and its modifications, by running chronologically
// ordered migrations
func Migrate(ctx context.Context, db *sql.DB, logger Logger) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}

	defer conn.Close()

	sources, err := iofs.New(migrations, "migration")
	if err != nil {
		return err
	}

	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{
		DatabaseName: "logs_buffer",
	})
	if err != nil {
		return err
	}

	migrator, err := migrate.NewWithInstance("iofs", sources, "traces", driver)
	if err != nil {
		return err
	}

	defer migrator.Close()

	currentVersion, dirty, err := migrator.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		currentVersion = 0
	} else if err != nil {
		return err
	}

	logger.InfoContext(ctx, "before migration",
		"version", currentVersion,
		"dirty", dirty,
	)

	if err = migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	currentVersion, dirty, err = migrator.Version()
	if err != nil {
		return err
	}

	logger.InfoContext(ctx, "after migration",
		"version", currentVersion,
		"dirty", dirty,
	)

	return nil
}
