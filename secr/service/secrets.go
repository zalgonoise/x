package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zalgonoise/x/errors"
	"github.com/zalgonoise/x/secr/bolt"
	"github.com/zalgonoise/x/secr/crypt"
	"github.com/zalgonoise/x/secr/keys"
	"github.com/zalgonoise/x/secr/secret"
	"github.com/zalgonoise/x/secr/shared"
	"github.com/zalgonoise/x/secr/sqlite"
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
		return errors.Join(ErrInvalidUser, err)
	}
	if isShared, err := secret.ValidateKey(key); err != nil || isShared {
		return errors.Join(ErrInvalidKey, err)
	}
	if err := secret.ValidateValue(value); err != nil {
		return errors.Join(ErrInvalidValue, err)
	}

	u, err := s.users.Get(ctx, username)
	if err != nil {
		return fmt.Errorf("failed to fetch user: %w", err)
	}

	tx := newTx()

	// check if secret already exists
	oldSecr, err := s.secrets.Get(ctx, username, key)
	if err != nil && !errors.Is(sqlite.ErrNotFoundSecret, err) {
		return fmt.Errorf("failed to fetch previous secret under this key: %w", err)
	}
	if oldSecr != nil {
		// remove old shares if they exist
		shares, err := s.shares.Get(ctx, username, oldSecr.Key)
		if err != nil && !errors.Is(sqlite.ErrNotFoundShare, err) {
			return fmt.Errorf("failed to fetch previous shared secrets: %w", err)
		}
		for _, sh := range shares {
			tx.Add(func() error {
				_, err := s.shares.Create(ctx, sh)
				return err
			})
			err := s.shares.Delete(ctx, sh)
			if err != nil {
				return tx.Rollback(fmt.Errorf("failed to remove old share: %w", err))
			}
		}

		// get encrypted value for existing secret (for RollbackFn)
		val, err := s.keys.Get(ctx, keys.UserBucket(u.ID), key)
		if err != nil {
			return tx.Rollback(fmt.Errorf("failed to fetch old secret's value: %w", err))
		}

		tx.Add(func() error {
			return s.keys.Set(ctx, keys.UserBucket(u.ID), key, val)
		})

		// remove it
		err = s.keys.Delete(ctx, keys.UserBucket(u.ID), key)
		if err != nil {
			return tx.Rollback(fmt.Errorf("failed to remove old secret's value: %w", err))
		}

		tx.Add(func() error {
			_, err := s.secrets.Create(ctx, username, oldSecr)
			return err
		})

		// remove the secret's metadata
		err = s.secrets.Delete(ctx, username, key)
		if err != nil {
			return tx.Rollback(fmt.Errorf("failed to remove old secret: %w", err))
		}
	}

	// encrypt secret with user's key:
	// fetch the key
	cipherKey, err := s.keys.Get(ctx, keys.UserBucket(u.ID), keys.UniqueID)
	if err != nil {
		return tx.Rollback(fmt.Errorf("failed to get user's private key: %w", err))
	}

	// encrypt value with user's private key
	cipher := crypt.NewCipher(cipherKey)
	encValue, err := cipher.Encrypt(value)
	if err != nil {
		return tx.Rollback(fmt.Errorf("failed to encrypt value: %w", err))
	}

	// store encrypted value
	err = s.keys.Set(ctx, keys.UserBucket(u.ID), key, encValue)
	if err != nil {
		return tx.Rollback(fmt.Errorf("failed to store the secret: %w", err))
	}
	tx.Add(func() error {
		if oldSecr == nil {
			return s.keys.Delete(ctx, keys.UserBucket(u.ID), key)
		}
		return nil
	})

	secr := &secret.Secret{
		Key: key,
	}

	// create secret
	id, err := s.secrets.Create(ctx, username, secr)
	if err != nil {
		return tx.Rollback(fmt.Errorf("failed to create the secret: %w", err))
	}
	secr.ID = id

	return nil
}

// GetSecret fetches the secret with key `key`, for user `username`. Returns a secret and an error
func (s service) GetSecret(ctx context.Context, username string, key string) (*secret.Secret, error) {
	if err := user.ValidateUsername(username); err != nil {
		return nil, errors.Join(ErrInvalidUser, err)
	}
	isShared, err := secret.ValidateKey(key)
	if err != nil {
		return nil, errors.Join(ErrInvalidKey, err)
	}

	// handle fetching an owned secret vs a shared secret
	if isShared {
		userAndKey := strings.Split(key, ":")
		return s.getSharedSecret(ctx, userAndKey[0], userAndKey[1], username)
	}

	return s.getSecret(ctx, username, key)
}

func (s service) getSecret(ctx context.Context, username, key string) (*secret.Secret, error) {
	u, err := s.users.Get(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	// fetch secret('s metadata )
	secr, err := s.secrets.Get(ctx, username, key)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the secret: %w", err)
	}

	// fetch user's private key to decode encrypted secret
	cipherKey, err := s.keys.Get(ctx, keys.UserBucket(u.ID), keys.UniqueID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the user's private key: %w", err)
	}
	cipher := crypt.NewCipher(cipherKey)

	// fetch secret's value
	encValue, err := s.keys.Get(ctx, keys.UserBucket(u.ID), key)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the secret: %w", err)
	}

	// decrypt value with user's private key
	decValue, err := cipher.Decrypt(encValue)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt value: %w", err)
	}

	secr.Value = string(decValue)
	return secr, nil
}

func (s service) getSharedSecret(ctx context.Context, owner, key, target string) (*secret.Secret, error) {
	// get the original share (as if it was the owner)
	sh, err := s.shares.Get(ctx, owner, key)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch shared secret metadata: %w", err)
	}

	var validShare *shared.Share
	for _, share := range sh {
		// validate share's deadline first
		if share.Until != nil {
			if time.Now().After(*share.Until) {
				// remove it if expired
				err := s.shares.Delete(ctx, share)
				if err != nil {
					return nil, fmt.Errorf("failed to remove expired shared secret: %w", err)
				}
				continue
			}
		}
		// check if the caller is one of the targets
		for _, t := range share.Target {
			if t == target {
				validShare = share
				break
			}
		}
	}

	// no share found, caller requests a share that doesn't exist
	// or doesn't have access to
	if validShare == nil {
		return nil, ErrZeroShares
	}

	// fetch the deciphered secret
	secr, err := s.getSecret(ctx, owner, key)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch shared secret: %w", err)
	}
	// erase creation time; set key as `user:key`
	secr.CreatedAt = time.Time{}
	secr.Key = fmt.Sprintf("%s:%s", owner, secr.Key)
	return secr, nil
}

// ListSecrets retuns all secrets for user `username`. Returns a list of secrets and an error
func (s service) ListSecrets(ctx context.Context, username string) ([]*secret.Secret, error) {
	if err := user.ValidateUsername(username); err != nil {
		return nil, errors.Join(ErrInvalidUser, err)
	}

	u, err := s.users.Get(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	// fetch secrets(' metadata )
	secrets, err := s.secrets.List(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to list the user's secrets: %w", err)
	}

	// fetch user's private key to decode encrypted secret
	cipherKey, err := s.keys.Get(ctx, keys.UserBucket(u.ID), keys.UniqueID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the user's private key: %w", err)
	}
	cipher := crypt.NewCipher(cipherKey)

	// fetch and decode the value for each secret
	for _, secr := range secrets {
		// fetch secret's value
		encValue, err := s.keys.Get(ctx, keys.UserBucket(u.ID), secr.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch the secret: %w", err)
		}

		// decrypt value with user's private key
		decValue, err := cipher.Decrypt(encValue)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt value: %w", err)
		}

		secr.Value = string(decValue)
	}

	// aggregate secrets that are shared with this user
	sharedSecrets, err := s.shares.ListTarget(ctx, username)
	if err != nil {
		return secrets, fmt.Errorf("failed to fetch secrets shared with %s: %w", username, err)
	}
	for _, sh := range sharedSecrets {
		// extract secret from shared secret
		sharedSecr, err := s.getSharedSecret(ctx, sh.Owner, sh.SecretKey, username)
		if err != nil {
			if errors.Is(ErrZeroShares, err) {
				continue
			}
			return secrets, fmt.Errorf("failed to fetch secrets shared with %s: %w", username, err)
		}
		// append results
		secrets = append(secrets, sharedSecr)
	}

	return secrets, nil
}

// DeleteSecret removes a secret with key `key` from the user `username`. Returns an error
func (s service) DeleteSecret(ctx context.Context, username string, key string) error {
	if err := user.ValidateUsername(username); err != nil {
		return errors.Join(ErrInvalidUser, err)
	}
	if isShared, err := secret.ValidateKey(key); err != nil || isShared {
		return errors.Join(ErrInvalidKey, err)
	}

	u, err := s.users.Get(ctx, username)
	if err != nil {
		return fmt.Errorf("failed to fetch user: %w", err)
	}

	tx := newTx()

	// remove shares for this secret
	shares, err := s.shares.Get(ctx, username, key)
	if err != nil && !errors.Is(sqlite.ErrNotFoundShare, err) {
		return fmt.Errorf("failed to list shared secrets: %w", err)
	}
	for _, sh := range shares {
		tx.Add(func() error {
			_, err := s.shares.Create(ctx, sh)
			return err
		})
		err := s.shares.Delete(ctx, sh)
		if err != nil {
			return tx.Rollback(fmt.Errorf("failed to remove shared secret: %w", err))
		}
	}

	// fetch the encrypted secret (for RollbackFn)
	secr, err := s.keys.Get(ctx, keys.UserBucket(u.ID), key)
	if err != nil {
		if errors.Is(bolt.ErrEmptyBucket, err) {
			// nothing to delete, no changes in state
			return nil
		}
		return tx.Rollback(fmt.Errorf("failed to fetch the secret: %w", err))
	}
	tx.Add(func() error {
		return s.keys.Set(ctx, keys.UserBucket(u.ID), key, secr)
	})

	// delete it
	err = s.keys.Delete(ctx, keys.UserBucket(u.ID), key)
	if err != nil {
		return tx.Rollback(fmt.Errorf("failed to remove secret: %w", err))
	}

	// fetch the secret's metadata (for RollbackFn)
	secretMeta, err := s.secrets.Get(ctx, username, key)
	if err != nil {
		return tx.Rollback(fmt.Errorf("failed to fetch secret: %w", err))
	}

	tx.Add(func() error {
		_, err := s.secrets.Create(ctx, username, secretMeta)
		return err
	})

	// delete it
	err = s.secrets.Delete(ctx, username, key)
	if err != nil {
		return tx.Rollback(fmt.Errorf("failed to remove secret: %w", err))
	}

	return nil
}
