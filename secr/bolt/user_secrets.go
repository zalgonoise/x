package bolt

import (
	"context"
	"fmt"

	"go.etcd.io/bbolt"

	"github.com/zalgonoise/x/secr/secret"
)

type secretRepository struct {
	db *bbolt.DB
}

func NewSecretStoreRepository(db *bbolt.DB) secret.Repository {
	return &secretRepository{db}
}

func (sr *secretRepository) Create(ctx context.Context, username string, s *secret.Secret) error {
	if username == "" {
		return ErrEmptyUsername
	}
	if s.Key == "" {
		return ErrEmptySecretKey
	}
	if len(s.Value) == 0 {
		return ErrEmptySecretValue
	}

	err := sr.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(username))
		if b == nil {
			return ErrEmptyBucket
		}

		err := b.Put([]byte(s.Key), s.Value)
		if err != nil {
			return fmt.Errorf("failed to create secret: %v", err)
		}

		return nil

	})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDBError, err)
	}
	return nil
}

func (sr *secretRepository) Get(ctx context.Context, username string, key string) (*secret.Secret, error) {
	if username == "" {
		return nil, ErrEmptyUsername
	}
	if key == "" {
		return nil, ErrEmptySecretKey
	}

	var s = new(secret.Secret)

	err := sr.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(username))
		if b == nil {
			return ErrNotFoundKey
		}
		s.Value = b.Get([]byte(key))

		if len(s.Value) == 0 {
			return fmt.Errorf("key %s: %v", key, ErrSecretUnset)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDBError, err)
	}

	return s, nil
}

func (sr *secretRepository) List(ctx context.Context, username string) ([]*secret.Secret, error) {
	if username == "" {
		return nil, ErrEmptyUsername
	}

	var s = []*secret.Secret{}

	err := sr.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(username))
		if b == nil {
			return ErrNotFoundKey
		}

		err := b.ForEach(func(k, v []byte) error {
			if string(k) == string(ident) {
				// skip user's unique key
				return nil
			}
			s = append(s, &secret.Secret{Key: string(k), Value: v})
			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to list secrets: %v", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDBError, err)
	}

	return s, nil
}
func (sr *secretRepository) Delete(ctx context.Context, username string, key string) error {
	if username == "" {
		return ErrEmptyUsername
	}
	if key == "" {
		return ErrEmptySecretKey
	}

	err := sr.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(username))
		if b == nil {
			return ErrEmptyBucket
		}

		err := b.Delete([]byte(key))
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
