package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/zalgonoise/x/secr/bolt"
	"github.com/zalgonoise/x/secr/crypt"
	"github.com/zalgonoise/x/secr/keys"
	"github.com/zalgonoise/x/secr/secret"
)

var (
	ErrNoKey   = errors.New("no secret key provided")
	ErrNoValue = errors.New("no secret value provided")
)

func (s service) CreateSecret(ctx context.Context, username string, key string, value []byte) error {
	if username == "" {
		return fmt.Errorf("%w: username cannot be empty", ErrNoUser)
	}
	if key == "" {
		return fmt.Errorf("%w: key cannot be empty", ErrNoKey)
	}
	if len(value) == 0 {
		return fmt.Errorf("%w: value cannot be empty", ErrNoValue)
	}

	// encrypt secret with user's key:
	// fetch the key
	cipherKey, err := s.keys.Get(ctx, username, keys.UniqueID)
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

	err = s.keys.Set(ctx, username, key, encValue)
	if err != nil {
		return fmt.Errorf("failed to store the secret: %v", err)
	}
	var rollback = func() error {
		return s.keys.Delete(ctx, username, key)
	}

	secr := &secret.Secret{
		Key: key,
	}
	err = s.secrets.Create(ctx, username, secr)
	if err != nil {
		rerr := rollback()
		if rerr != nil {
			err = fmt.Errorf("%w -- rollback error: %v", err, rerr)
		}
		return fmt.Errorf("failed to create the secret: %v", err)
	}
	return nil
}
func (s service) GetSecret(ctx context.Context, username string, key string) (*secret.Secret, error) {
	if username == "" {
		return nil, fmt.Errorf("%w: username cannot be empty", ErrNoUser)
	}
	if key == "" {
		return nil, fmt.Errorf("%w: key cannot be empty", ErrNoKey)
	}

	// fetch secret('s metadata )
	secr, err := s.secrets.Get(ctx, username, key)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the secret: %v", err)
	}

	// fetch user's private key to decode encrypted secret
	cipherKey, err := s.keys.Get(ctx, username, keys.UniqueID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the user's private key: %v", err)
	}
	cipher, err := crypt.NewCipher(cipherKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create a cipher from the user's private key: %v", err)
	}

	// fetch secret's value
	encValue, err := s.keys.Get(ctx, username, key)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the secret: %v", err)
	}

	// decrypt value with user's private key
	decValue := make([]byte, 0, len(encValue))
	cipher.Decrypt(decValue, encValue)

	secr.Value = decValue
	return secr, nil
}
func (s service) ListSecrets(ctx context.Context, username string) ([]*secret.Secret, error) {
	if username == "" {
		return nil, fmt.Errorf("%w: username cannot be empty", ErrNoUser)
	}

	// fetch secret('s metadata )
	secrets, err := s.secrets.List(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to list the user's secrets: %v", err)
	}

	// fetch user's private key to decode encrypted secret
	cipherKey, err := s.keys.Get(ctx, username, keys.UniqueID)
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
		encValue, err := s.keys.Get(ctx, username, secr.Key)
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
func (s service) DeleteSecret(ctx context.Context, username string, key string) error {
	if username == "" {
		return fmt.Errorf("%w: username cannot be empty", ErrNoUser)
	}
	if key == "" {
		return fmt.Errorf("%w: key cannot be empty", ErrNoKey)
	}

	secr, err := s.keys.Get(ctx, username, key)
	if err != nil {
		if errors.Is(bolt.ErrEmptyBucket, err) {
			// nothing to delete, no changes in state
			return nil
		}
		return fmt.Errorf("failed to fetch the secret: %v", err)
	}
	var rollback = func() error {
		return s.keys.Set(ctx, username, key, secr)
	}

	err = s.keys.Delete(ctx, username, key)
	if err != nil {
		rerr := rollback()
		if rerr != nil {
			err = fmt.Errorf("%w -- rollback error: %v", err, rerr)
		}
		return err
	}
	return nil
}
