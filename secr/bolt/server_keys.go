package bolt

import (
	"context"
	"fmt"

	"go.etcd.io/bbolt"
)

var serverID = []byte("secr-server")

type ServerKeysRepository struct {
	db *bbolt.DB
}

func NewServerRepository(db *bbolt.DB) *ServerKeysRepository {
	return &ServerKeysRepository{db}
}

func (skr *ServerKeysRepository) Get(ctx context.Context, key string) ([]byte, error) {
	if key == "" {
		return nil, ErrEmptySecretKey
	}

	var k = make([]byte, 0, 128)
	err := skr.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(serverID)
		if b == nil {
			return ErrEmptyBucket
		}
		k = b.Get([]byte(key))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDBError, err)
	}
	return k, nil

}
func (skr *ServerKeysRepository) Set(ctx context.Context, key string, value []byte) error {
	if key == "" {
		return ErrEmptySecretKey
	}
	if len(value) == 0 {
		return ErrEmptySecretValue
	}
	err := skr.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(serverID)
		if err != nil {
			return fmt.Errorf("failed to create bucket for the server: %v", err)
		}

		err = b.Put([]byte(key), value)
		if err != nil {
			return fmt.Errorf("failed to set value in the server's bucket: %v", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDBError, err)
	}
	return nil
}
