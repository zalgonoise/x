package database

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

// Connect creates a new instance of a SQLite database from the input URI. This URI can be empty (or ":memory:", if the
// caller wishes to run the database in-memory.
//
// TODO: implement database schema; implement database migration logic
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
