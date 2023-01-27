package factory

import (
	"os"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zalgonoise/x/secr/bolt"
	"github.com/zalgonoise/x/secr/keys"
	"github.com/zalgonoise/x/secr/secret"
	"github.com/zalgonoise/x/secr/shared"
	"github.com/zalgonoise/x/secr/sqlite"
	"github.com/zalgonoise/x/secr/user"
)

const (
	sqliteDbPath = "/secr/sqlite.db"
	boltDbPath   = "/secr/keys.db"
)

// SQLite creates user and secret repositories based on the defined SQLite DB path
func SQLite(path string) (user.Repository, secret.Repository, shared.Repository, error) {
	fs, err := os.Stat(path)
	if (err != nil && os.IsNotExist(err)) || (fs != nil && fs.Size() == 0) {
		_, err := os.Create(path)
		if err != nil {
			if path == sqliteDbPath {
				return nil, nil, nil, err
			}
			return SQLite(sqliteDbPath)
		}
	}

	db, err := sqlite.Open(path)
	if err != nil {
		if path == sqliteDbPath {
			return nil, nil, nil, err
		}
		return SQLite(sqliteDbPath)
	}

	return user.WithTrace(sqlite.NewUserRepository(db)),
		secret.WithTrace(sqlite.NewSecretRepository(db)),
		shared.WithTrace(sqlite.NewSharedRepository(db)),
		nil
}

// Bolt creates a key repository based on the defined Bolt DB path
func Bolt(path string) (keys.Repository, error) {
	fs, err := os.Stat(path)
	if (err != nil && os.IsNotExist(err)) || (fs != nil && fs.Size() == 0) {
		_, err := os.Create(path)
		if err != nil {
			if path == boltDbPath {
				return nil, err
			}
			return Bolt(boltDbPath)
		}
	}

	db, err := bolt.Open(path)
	if err != nil {
		if path == boltDbPath {
			return nil, err
		}
		return Bolt(boltDbPath)
	}
	return keys.WithTrace(bolt.NewKeysRepository(db)), nil
}
