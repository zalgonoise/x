package bolt

import (
	"go.etcd.io/bbolt"
)

// Open will initialize a Bolt DB based on the `.db` file in `path`,
// returning a pointer to a bbolt.DB and an error
func Open(path string) (*bbolt.DB, error) {
	return bbolt.Open(path, 0600, nil)
}
