package service

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/zalgonoise/x/errors"
	"github.com/zalgonoise/x/secr/crypt"
	"github.com/zalgonoise/x/secr/keys"
	"github.com/zalgonoise/x/secr/sqlite"
	"github.com/zalgonoise/x/secr/user"
)

var (
	ErrInvalidUser       = errors.New("error validating user")
	ErrInvalidPassword   = errors.New("error validating password")
	ErrInvalidName       = errors.New("error validating name")
	ErrAlreadyExistsUser = errors.New("user already exists")
)

// CreateUser creates the user under username `username`, with the provided password `password` and name `name`
// It returns a user and an error
func (s service) CreateUser(ctx context.Context, username, password, name string) (*user.User, error) {
	if err := user.ValidateUsername(username); err != nil {
		return nil, errors.Join(ErrInvalidUser, err)
	}
	if err := user.ValidatePassword(password); err != nil {
		return nil, errors.Join(ErrInvalidPassword, err)
	}
	if err := user.ValidateName(name); err != nil {
		return nil, errors.Join(ErrInvalidName, err)
	}

	// check if user exists
	_, err := s.users.Get(ctx, username)
	if err == nil || !errors.Is(sqlite.ErrNotFoundUser, err) {
		return nil, fmt.Errorf("failed to create user: %w", ErrAlreadyExistsUser)
	}

	// generate hash from password
	salt := crypt.NewSalt()
	hashedPassword := crypt.Hash([]byte(password), salt[:])

	encSalt := base64.StdEncoding.EncodeToString(salt[:])
	encHash := base64.StdEncoding.EncodeToString(hashedPassword[:])

	// create the user
	u := &user.User{
		Username: username,
		Hash:     encHash,
		Salt:     encSalt,
		Name:     name,
	}
	id, err := s.users.Create(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to create user %s: %w", username, err)
	}
	u.ID = id

	tx := newTx()
	tx.Add(func() error {
		return s.users.Delete(ctx, username)
	})

	// generate a new private key for this user, and store it
	key := crypt.New32Key()
	err = s.keys.Set(ctx, keys.UserBucket(u.ID), keys.UniqueID, key[:])
	if err != nil {
		return nil, tx.Rollback(fmt.Errorf("failed to create user %s: %w", username, err))
	}

	return u, nil
}

// GetUser fetches the user with username `username`. Returns a user and an error
func (s service) GetUser(ctx context.Context, username string) (*user.User, error) {
	if err := user.ValidateUsername(username); err != nil {
		return nil, errors.Join(ErrInvalidUser, err)
	}

	u, err := s.users.Get(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user %s: %w", username, err)
	}
	return u, nil
}

// ListUsers returns all the users in the directory, and an error
func (s service) ListUsers(ctx context.Context) ([]*user.User, error) {
	users, err := s.users.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

// UpdateUser updates the user `username`'s name, found in `updated` user. Returns an error
func (s service) UpdateUser(ctx context.Context, username string, updated *user.User) error {
	if err := user.ValidateUsername(username); err != nil {
		return errors.Join(ErrInvalidUser, err)
	}
	if updated == nil {
		return fmt.Errorf("%w: updated user cannot be nil", ErrInvalidUser)
	}
	if err := user.ValidateName(updated.Name); err != nil {
		return errors.Join(ErrInvalidName, err)
	}

	u, err := s.users.Get(ctx, username)
	if err != nil {
		return fmt.Errorf("failed to fetch original user %s: %w", username, err)
	}

	if updated.Name == u.Name {
		// no changes to be made
		return nil
	}
	u.Name = updated.Name

	err = s.users.Update(ctx, username, u)
	if err != nil {
		return fmt.Errorf("failed to update user %s: %w", username, err)
	}

	return nil
}

// DeleteUser removes the user with username `username`. Returns an error
func (s service) DeleteUser(ctx context.Context, username string) error {
	if err := user.ValidateUsername(username); err != nil {
		return errors.Join(ErrInvalidUser, err)
	}

	u, err := s.users.Get(ctx, username)
	if err != nil {
		if errors.Is(sqlite.ErrNotFoundUser, err) {
			// no change in state
			return nil
		}
		return fmt.Errorf("failed to fetch original user %s: %w", username, err)
	}

	tx := newTx()

	// remove all shares
	shares, err := s.shares.List(ctx, username)
	if err != nil {
		return fmt.Errorf("failed to fetch shared secrets: %w", err)
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

	// remove all secrets
	secrets, err := s.secrets.List(ctx, username)
	if err != nil {
		return fmt.Errorf("failed to list user secrets: %w", err)
	}
	for _, secr := range secrets {
		tx.Add(func() error {
			_, err := s.secrets.Create(ctx, username, secr)
			return err
		})

		err := s.secrets.Delete(ctx, username, secr.Key)
		if err != nil {
			return tx.Rollback(fmt.Errorf("failed to remove secret: %w", err))
		}
	}

	// get user's private key for rollback func
	upk, err := s.keys.Get(ctx, keys.UserBucket(u.ID), keys.UniqueID)
	if err != nil {
		return fmt.Errorf("failed to fetch user %s's key: %w", username, err)
	}
	tx.Add(func() error {
		return s.keys.Set(ctx, keys.UserBucket(u.ID), keys.UniqueID, upk)
	})

	// delete private key
	err = s.keys.Delete(ctx, keys.UserBucket(u.ID), keys.UniqueID)
	if err != nil {
		return tx.Rollback(fmt.Errorf("failed to delete user %s's key: %w", username, err))
	}

	// delete user
	err = s.users.Delete(ctx, username)
	if err != nil {
		return tx.Rollback(fmt.Errorf("failed to delete user %s: %w", username, err))
	}

	return nil
}
