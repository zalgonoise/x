package database

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite" // Database driver
)

const (
	uriFormat = "file:%s?cache=shared"
	inMemory  = ":memory:"
	minAlloc  = 64
)

var (
	//go:embed migration/ca/*
	caMigrationFiles embed.FS

	//go:embed migration/authz/*
	authzMigrationFiles embed.FS
)

var ErrInvalidServiceType = errors.New("invalid service type")

type Service string

const (
	AuthzService Service = "authz"
	CAService    Service = "certificate_authority"
)

func Open(uri string) (*sql.DB, error) {
	switch uri {
	case inMemory:
	case "":
		uri = inMemory
	default:
		if err := validateURI(uri); err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("sqlite", fmt.Sprintf(uriFormat, uri))
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

func Migrate(ctx context.Context, db *sql.DB, log *slog.Logger, service Service) error {
	var migration fs.FS

	switch service {
	case AuthzService:
		migration = authzMigrationFiles
	case CAService:
		migration = caMigrationFiles
	default:
		return fmt.Errorf("%w: %q", ErrInvalidServiceType, service)
	}

	return runMigrations(ctx, db, log, migration)
}

func runMigrations(ctx context.Context, db *sql.DB, log *slog.Logger, migrationFiles fs.FS) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}

	defer conn.Close()

	source, err := iofs.New(migrationFiles, "migration")
	if err != nil {
		return err
	}

	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return err
	}

	migrator, err := migrate.NewWithInstance("iofs", source, "timesheet", driver)
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

	log.InfoContext(ctx, "before migration", slog.Uint64("version", uint64(currentVersion)), slog.Bool("dirty", dirty))

	err = migrator.Up()
	if !errors.Is(err, migrate.ErrNoChange) && err != nil {
		return err
	}

	currentVersion, dirty, err = migrator.Version()
	if err != nil {
		return err
	}

	log.InfoContext(ctx, "after migration", slog.Uint64("version", uint64(currentVersion)), slog.Bool("dirty", dirty))

	return nil
}
