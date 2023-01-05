package bolt

import (
	"context"
	"errors"
	"fmt"

	"github.com/zalgonoise/x/secr/keys"
	"go.etcd.io/bbolt"
)

var (
	ErrDBError     = errors.New("database error")
	ErrNotFoundKey = errors.New("couldn't find the key")
	ErrEmptyKey    = errors.New("key cannot be empty")
	ErrEmptyValue  = errors.New("username cannot be empty")
	ErrEmptyBucket = errors.New("empty bucket")
	ErrForbidden   = errors.New("unable to modify this resource")
)

type keysRepository struct {
	db *bbolt.DB
}

func NewKeysRepository(db *bbolt.DB) keys.Repository {
	return &keysRepository{db}
}

func (ukr *keysRepository) Get(ctx context.Context, bucket, k string) ([]byte, error) {
	if bucket == "" {
		return nil, ErrEmptyBucket
	}
	if k == "" {
		return nil, ErrEmptyKey
	}

	var v []byte

	err := ukr.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return ErrEmptyBucket
		}
		v = b.Get([]byte(k))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDBError, err)
	}

	return v, nil
}

func (ukr *keysRepository) Set(ctx context.Context, bucket, k string, v []byte) error {
	if bucket == "" {
		return ErrEmptyBucket
	}
	if k == "" {
		return ErrEmptyKey
	}
	if len(k) == 0 {
		return ErrEmptyValue
	}

	err := ukr.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return fmt.Errorf("failed to get / create bucket: %v", err)
		}

		err = b.Put([]byte(k), []byte(v))
		if err != nil {
			return fmt.Errorf("failed to set key-value: %v", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDBError, err)
	}
	return nil
}

func (ukr *keysRepository) Delete(ctx context.Context, bucket, k string) error {
	if bucket == "" {
		return ErrEmptyBucket
	}
	if k == "" {
		return ErrEmptyKey
	}

	err := ukr.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return ErrEmptyBucket
		}

		err := b.Delete([]byte(k))
		if err != nil {
			return fmt.Errorf("failed to delete key %s in the bucket %s: %v", k, bucket, err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("%w: %v", ErrDBError, err)
	}
	return nil
}

func (ukr *keysRepository) Purge(ctx context.Context, bucket string) error {
	if bucket == "" {
		return ErrEmptyBucket
	}
	if bucket == keys.ServerID {
		return ErrForbidden
	}

	err := ukr.db.Update(func(tx *bbolt.Tx) error {
		err := tx.DeleteBucket([]byte(bucket))
		if err != nil {
			return fmt.Errorf("failed to delete the bucket %s: %v", bucket, err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("%w: %v", ErrDBError, err)
	}
	return nil

}
