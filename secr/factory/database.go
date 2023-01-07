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
)

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

func Bolt() (keys.Repository, error) {
	fs, err := os.Stat(sqliteDbPath)
	if (err != nil && os.IsNotExist(err)) || (fs != nil && fs.Size() == 0) {
		_, err := os.Create(sqliteDbPath)
		if err != nil {
			return nil, err
		}
	}

	db, err := bolt.Open(sqliteDbPath)
	if err != nil {
		return nil, err
	}
	return bolt.NewKeysRepository(db), nil

}
