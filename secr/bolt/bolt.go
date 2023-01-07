package bolt

import (
	"go.etcd.io/bbolt"
)

func Open(path string) (*bbolt.DB, error) {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	return db, nil
}
