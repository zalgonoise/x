package bolt

import (
	"context"
	"errors"
	"fmt"

	"github.com/zalgonoise/x/secr/keys"
	"go.etcd.io/bbolt"
)

var (
	ident = []byte("unique_key")

	ErrDBError       = errors.New("database error")
	ErrNotFoundKey   = errors.New("couldn't find the key")
	ErrEmptyKey      = errors.New("key cannot be empty")
	ErrEmptyUsername = errors.New("username cannot be empty")
	ErrEmptyBucket   = errors.New("user's secrets must be initialized with the user's unique key")
)

type userKeysRepository struct {
	db *bbolt.DB
}

func NewUserKeysRepository(db *bbolt.DB) keys.Repository {
	return &userKeysRepository{db}
}

func (ukr *userKeysRepository) Get(ctx context.Context, username string) (keys.Key, error) {
	if username == "" {
		return nil, ErrEmptyUsername
	}

	var k keys.Key

	err := ukr.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(username))
		if b == nil {
			return ErrEmptyBucket
		}
		k = b.Get(ident)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDBError, err)
	}

	return k, nil
}

func (ukr *userKeysRepository) Set(ctx context.Context, username string, k keys.Key) error {
	if username == "" {
		return ErrEmptyUsername
	}
	if len(k) == 0 {
		return ErrEmptyKey
	}

	err := ukr.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(username))
		if err != nil {
			return fmt.Errorf("failed to create bucket for the user: %v", err)
		}

		err = b.Put(ident, []byte(k))
		if err != nil {
			return fmt.Errorf("failed to set value in the user's bucket: %v", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDBError, err)
	}
	return nil
}

func (ukr *userKeysRepository) Delete(ctx context.Context, username string) error {
	if username == "" {
		return ErrEmptyUsername
	}

	err := ukr.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(username))
		if b == nil {
			return ErrEmptyBucket
		}

		err := b.Delete(ident)
		if err != nil {
			return fmt.Errorf("failed to delete key in the user's bucket: %v", err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("%w: %v", ErrDBError, err)
	}
	return nil
}
