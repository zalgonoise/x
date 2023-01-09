package factory

import (
	"os"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zalgonoise/x/secr/bolt"
	"github.com/zalgonoise/x/secr/keys"
	"github.com/zalgonoise/x/secr/secret"
	"github.com/zalgonoise/x/secr/sqlite"
	"github.com/zalgonoise/x/secr/user"
)

const (
	sqliteDbPath = "/secr/sqlite.db"
	boltDbPath   = "/secr/keys.db"
)

// SQLite creates user and secret repositories based on the defined SQLite DB path
func SQLite() (user.Repository, secret.Repository, error) {
	fs, err := os.Stat(sqliteDbPath)
	if (err != nil && os.IsNotExist(err)) || (fs != nil && fs.Size() == 0) {
		_, err := os.Create(sqliteDbPath)
		if err != nil {
			return nil, nil, err
		}
	}

	db, err := sqlite.Open(sqliteDbPath)
	if err != nil {
		return nil, nil, err
	}

	return sqlite.NewUserRepository(db), sqlite.NewSecretRepository(db), nil
}

// Bolt creates a key repository based on the defined Bolt DB path
func Bolt() (keys.Repository, error) {
	fs, err := os.Stat(boltDbPath)
	if (err != nil && os.IsNotExist(err)) || (fs != nil && fs.Size() == 0) {
		_, err := os.Create(boltDbPath)
		if err != nil {
			return nil, err
		}
	}

	db, err := bolt.Open(boltDbPath)
	if err != nil {
		return nil, err
	}
	return bolt.NewKeysRepository(db), nil

}
