package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zalgonoise/x/secr/bolt"
	"github.com/zalgonoise/x/secr/crypt"
	"github.com/zalgonoise/x/secr/keys"
	"github.com/zalgonoise/x/secr/secret"
	"github.com/zalgonoise/x/secr/shared"
	"github.com/zalgonoise/x/secr/user"
)

var (
	ErrInvalidKey   = errors.New("error validating secret key")
	ErrInvalidValue = errors.New("error validating secret value")
	ErrZeroShares   = errors.New("no shared secrets found with the input key")
)

// CreateSecret creates a secret with key `key` and value `value` (as a slice of bytes), for the
// user `username`. It returns an error
func (s service) CreateSecret(ctx context.Context, username string, key string, value []byte) error {
	if err := user.ValidateUsername(username); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidUser, err)
	}
	if isShared, err := secret.ValidateKey(key); err != nil || isShared {
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

	cipher := crypt.NewCipher(cipherKey)
	// encrypt value with user's private key
	encValue, err := cipher.Encrypt(value)
	if err != nil {
		return fmt.Errorf("failed to encrypt value: %v", err)
	}

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
	isShared, err := secret.ValidateKey(key)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidKey, err)
	}

	if isShared {
		userAndKey := strings.Split(key, ":")
		return s.getSharedSecret(ctx, userAndKey[0], userAndKey[1], username)
	}

	return s.getSecret(ctx, username, key)
}

func (s service) getSecret(ctx context.Context, username, key string) (*secret.Secret, error) {
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
	cipher := crypt.NewCipher(cipherKey)

	// fetch secret's value
	encValue, err := s.keys.Get(ctx, keys.UserBucket(u.ID), key)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the secret: %v", err)
	}

	// decrypt value with user's private key
	decValue, err := cipher.Decrypt(encValue)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt value: %v", err)
	}

	secr.Value = string(decValue)
	return secr, nil
}

func (s service) getSharedSecret(ctx context.Context, owner, key, target string) (*secret.Secret, error) {
	sh, err := s.shares.Get(ctx, owner, key)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch shared secret metadata: %v", err)
	}

	var validShare *shared.Share

	for _, share := range sh {
		if share.Until != nil {
			if time.Now().After(*share.Until) {
				err := s.shares.Delete(ctx, share)
				if err != nil {
					return nil, fmt.Errorf("failed to remove expired shared secret: %v", err)
				}
				continue
			}
		}
		for _, t := range share.Target {
			if t.Username == target {
				validShare = share
				break
			}
		}
	}

	if validShare == nil {
		return nil, ErrZeroShares
	}

	secr, err := s.getSecret(ctx, owner, key)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch shared secret: %v", err)
	}
	secr.CreatedAt = time.Time{}
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
	cipher := crypt.NewCipher(cipherKey)

	// fetch and decode the value for each secret
	for _, secr := range secrets {
		// fetch secret's value
		encValue, err := s.keys.Get(ctx, keys.UserBucket(u.ID), secr.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch the secret: %v", err)
		}

		// decrypt value with user's private key
		decValue, err := cipher.Decrypt(encValue)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt value: %v", err)
		}

		secr.Value = string(decValue)
	}

	sharedSecrets, err := s.shares.ListTarget(ctx, username)
	if err != nil {
		return secrets, fmt.Errorf("failed to fetch secrets shared with %s: %v", username, err)
	}
	for _, sh := range sharedSecrets {
		sharedSecr, err := s.getSharedSecret(ctx, sh.Owner.Username, sh.Secret.Key, username)
		if err != nil {
			return secrets, fmt.Errorf("failed to fetch secrets shared with %s: %v", username, err)
		}
		secrets = append(secrets, sharedSecr)
	}

	return secrets, nil
}

// DeleteSecret removes a secret with key `key` from the user `username`. Returns an error
func (s service) DeleteSecret(ctx context.Context, username string, key string) error {
	if err := user.ValidateUsername(username); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidUser, err)
	}
	if isShared, err := secret.ValidateKey(key); err != nil || isShared {
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

	shares, err := s.shares.Get(ctx, username, key)
	if err != nil {
		return fmt.Errorf("failed to scan for shared secrets under this key: %v", err)
	}
	for _, sh := range shares {
		err = s.shares.Delete(ctx, sh)
		if err != nil {
			return fmt.Errorf("failed to remove shared secret: %v", err)
		}
	}
	return nil
}
