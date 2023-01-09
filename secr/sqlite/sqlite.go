package sqlite

import (
	"database/sql"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/1672703190_initial_up.sql
var initialMigration string

// Open will initialize a SQLite DB based on the `.sql` file in `path`,
// returning a pointer to a sql.DB and an error
//
// It executes the initial migration, as well.
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(initialMigration)
	if err != nil {
		return nil, err
	}
	return db, nil
}
