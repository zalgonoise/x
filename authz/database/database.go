package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/XSAM/otelsql"
	"github.com/zalgonoise/cfg"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	_ "modernc.org/sqlite" // Database driver
)

const (
	uriFormat = "file:%s?cache=shared"
	inMemory  = ":memory:"

	defaultMaxOpenConns = 16
	defaultMaxIdleConns = 16

	checkCATableExists = `
	SELECT EXISTS(SELECT 1 FROM sqlite_master
	WHERE type='table'
	AND name='services');
`

	createCATables = `
CREATE TABLE services
(
    id           INTEGER PRIMARY KEY NOT NULL,
    name         TEXT                NOT NULL UNIQUE,
    pub_key      BLOB                NOT NULL,
    cert         BLOB                NULL,
    expiry 			 DATETIME 					 NULL
);
`
)

var ErrInvalidServiceType = errors.New("invalid service type")

type Service string

const (
	AuthzService Service = "authz"
	CAService    Service = "certificate_authority"
)

func Open(uri string, opts ...cfg.Option[Config]) (*sql.DB, error) {
	switch uri {
	case inMemory:
	case "":
		uri = inMemory
	default:
		if err := validateURI(uri); err != nil {
			return nil, err
		}
	}

	config := cfg.New(opts...)

	db, err := otelsql.Open("sqlite",
		fmt.Sprintf(uriFormat, uri),
		otelsql.WithAttributes(semconv.DBSystemSqlite),
		otelsql.WithSpanOptions(otelsql.SpanOptions{
			OmitConnResetSession: true,
		}),
	)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(config.maxOpenConns)
	db.SetMaxIdleConns(config.maxIdleConns)

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

func Migrate(ctx context.Context, db *sql.DB, service Service, log *slog.Logger) error {
	switch service {
	case AuthzService:
		return runMigrations(ctx, db)
	case CAService:
		return runMigrations(ctx, db, migration{checkCATableExists, createCATables})
	default:
		return fmt.Errorf("%w: %q", ErrInvalidServiceType, service)
	}
}

type migration struct {
	check  string
	create string
}

func runMigrations(ctx context.Context, db *sql.DB, migrations ...migration) error {
	for i := range migrations {
		r, err := db.QueryContext(ctx, migrations[i].check)
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

		_, err = db.ExecContext(ctx, migrations[i].create)
		if err != nil {
			return err
		}
	}

	return nil
}
