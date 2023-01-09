package bolt

import (
	"go.etcd.io/bbolt"
)

// Open will initialize a Bolt DB based on the `.db` file in `path`,
// returning a pointer to a bbolt.DB and an error
func Open(path string) (*bbolt.DB, error) {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	return db, nil
}
