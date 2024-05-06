package database

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*
var migrations embed.FS

func Migrate(ctx context.Context, db *sql.DB, logger *slog.Logger) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		return err
	}

	driver, err := postgres.WithConnection(ctx, conn, &postgres.Config{})
	if err != nil {
		return err
	}

	migrator, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return err
	}

	defer migrator.Close()

	version, dirty, err := migrator.Version()

	if errors.Is(err, migrate.ErrNilVersion) {
		version = 0
	} else if err != nil {
		return err
	}

	logger.InfoContext(ctx, "before migration",
		slog.Int("version", int(version)),
		slog.Bool("dirty", dirty),
	)

	err = migrator.Up()
	if !errors.Is(err, migrate.ErrNoChange) && err != nil {
		return err
	}

	version, dirty, err = migrator.Version()
	if err != nil {
		return err
	}

	logger.InfoContext(ctx, "after migration",
		slog.Int("version", int(version)),
		slog.Bool("dirty", dirty),
	)

	return nil
}
