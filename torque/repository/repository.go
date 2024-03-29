package repository

import (
	"database/sql"
	"log/slog"
)

type SQLite struct {
	db *sql.DB

	logger *slog.Logger
}

func NewSQLite(db *sql.DB, logger *slog.Logger) *SQLite {
	return &SQLite{
		db:     db,
		logger: logger,
	}
}
