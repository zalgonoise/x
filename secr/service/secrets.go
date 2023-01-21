package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/zalgonoise/x/secr/bolt"
	"github.com/zalgonoise/x/secr/crypt"
	"github.com/zalgonoise/x/secr/keys"
	"github.com/zalgonoise/x/secr/secret"
	"github.com/zalgonoise/x/secr/user"
)

var (
	ErrInvalidKey   = errors.New("error validating secret key")
	ErrInvalidValue = errors.New("error validating secret value")
)

// CreateSecret creates a secret with key `key` and value `value` (as a slice of bytes), for the
// user `username`. It returns an error
func (s service) CreateSecret(ctx context.Context, username string, key string, value []byte) error {
	if err := user.ValidateUsername(username); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidUser, err)
	}
	if err := secret.ValidateKey(key); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidKey, err)
	}
	if err := secret.ValidateValue(value); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidValue, err)
	}

	u, err := s.GetUser(ctx, username)
	if err != nil {
		return fmt.Errorf("failed to fetch user: %v", err)
	}

	// encrypt secret with user's key:
	// fetch the key
	cipherKey, err := s.keys.Get(ctx, keys.UserBucket(u.ID), keys.UniqueID)
	if err != nil {
		return fmt.Errorf("failed to get user's private key: %v", err)
	}

	cipher, err := crypt.NewCipher(cipherKey)
	if err != nil {
		return fmt.Errorf("failed to create a cipher from the user's private key: %v", err)
	}

	// encrypt value with user's private key
	encValue := make([]byte, 0, len(value))
	cipher.Encrypt(encValue, value)

	err = s.keys.Set(ctx, keys.UserBucket(u.ID), key, encValue)
	if err != nil {
		return fmt.Errorf("failed to store the secret: %v", err)
	}
	var rollback = func() error {
		return s.keys.Delete(ctx, keys.UserBucket(u.ID), key)
	}

	secr := &secret.Secret{
		Key: key,
	}
	id, err := s.secrets.Create(ctx, username, secr)
	if err != nil {
		rerr := rollback()
		if rerr != nil {
			err = fmt.Errorf("%w -- rollback error: %v", err, rerr)
		}
		return fmt.Errorf("failed to create the secret: %v", err)
	}
	secr.ID = id

	return nil
}

// GetSecret fetches the secret with key `key`, for user `username`. Returns a secret and an error
func (s service) GetSecret(ctx context.Context, username string, key string) (*secret.Secret, error) {
	if err := user.ValidateUsername(username); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidUser, err)
	}
	if err := secret.ValidateKey(key); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidKey, err)
	}

	u, err := s.GetUser(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %v", err)
	}

	// fetch secret('s metadata )
	secr, err := s.secrets.Get(ctx, username, key)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the secret: %v", err)
	}

	// fetch user's private key to decode encrypted secret
	cipherKey, err := s.keys.Get(ctx, keys.UserBucket(u.ID), keys.UniqueID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the user's private key: %v", err)
	}
	cipher, err := crypt.NewCipher(cipherKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create a cipher from the user's private key: %v", err)
	}

	// fetch secret's value
	encValue, err := s.keys.Get(ctx, keys.UserBucket(u.ID), key)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the secret: %v", err)
	}

	// decrypt value with user's private key
	decValue := make([]byte, 0, len(encValue))
	cipher.Decrypt(decValue, encValue)

	secr.Value = decValue
	return secr, nil
}

// ListSecrets retuns all secrets for user `username`. Returns a list of secrets and an error
func (s service) ListSecrets(ctx context.Context, username string) ([]*secret.Secret, error) {
	if err := user.ValidateUsername(username); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidUser, err)
	}

	u, err := s.GetUser(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %v", err)
	}

	// fetch secret('s metadata )
	secrets, err := s.secrets.List(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to list the user's secrets: %v", err)
	}

	// fetch user's private key to decode encrypted secret
	cipherKey, err := s.keys.Get(ctx, keys.UserBucket(u.ID), keys.UniqueID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the user's private key: %v", err)
	}
	cipher, err := crypt.NewCipher(cipherKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create a cipher from the user's private key: %v", err)
	}

	// fetch and decode the value for each secret
	for _, secr := range secrets {
		// fetch secret's value
		encValue, err := s.keys.Get(ctx, keys.UserBucket(u.ID), secr.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch the secret: %v", err)
		}

		// decrypt value with user's private key
		decValue := make([]byte, 0, len(encValue))
		cipher.Decrypt(decValue, encValue)

		secr.Value = decValue
	}

	return secrets, nil
}

// DeleteSecret removes a secret with key `key` from the user `username`. Returns an error
func (s service) DeleteSecret(ctx context.Context, username string, key string) error {
	if err := user.ValidateUsername(username); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidUser, err)
	}
	if err := secret.ValidateKey(key); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidKey, err)
	}

	u, err := s.GetUser(ctx, username)
	if err != nil {
		return fmt.Errorf("failed to fetch user: %v", err)
	}

	secr, err := s.keys.Get(ctx, keys.UserBucket(u.ID), key)
	if err != nil {
		if errors.Is(bolt.ErrEmptyBucket, err) {
			// nothing to delete, no changes in state
			return nil
		}
		return fmt.Errorf("failed to fetch the secret: %v", err)
	}
	var rollback = func() error {
		return s.keys.Set(ctx, keys.UserBucket(u.ID), key, secr)
	}

	err = s.keys.Delete(ctx, keys.UserBucket(u.ID), key)
	if err != nil {
		rerr := rollback()
		if rerr != nil {
			err = fmt.Errorf("%w -- rollback error: %v", err, rerr)
		}
		return err
	}
	return nil
}
