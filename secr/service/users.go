package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/zalgonoise/x/secr/crypt"
	"github.com/zalgonoise/x/secr/keys"
	"github.com/zalgonoise/x/secr/sqlite"
	"github.com/zalgonoise/x/secr/user"
)

var (
	ErrNoUser            = errors.New("no user provided")
	ErrNoPassword        = errors.New("no password provided")
	ErrNoName            = errors.New("no name provided")
	ErrAlreadyExistsUser = errors.New("user already exists")
)

func (s service) CreateUser(ctx context.Context, username, password, name string) (*user.User, error) {
	if username == "" {
		return nil, fmt.Errorf("%w: username cannot be empty", ErrNoUser)
	}
	if password == "" {
		return nil, fmt.Errorf("%w: password cannot be empty", ErrNoPassword)
	}
	if name == "" {
		return nil, fmt.Errorf("%w: name cannot be empty", ErrNoName)
	}

	// check if user exists
	_, err := s.users.Get(ctx, username)
	if err == nil {
		return nil, fmt.Errorf("failed to create user: %v", ErrAlreadyExistsUser)
	}
	if !errors.Is(sqlite.ErrNotFoundUser, err) {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	// generate hash from password
	salt := crypt.NewSalt()
	hashedPassword := sha256.Sum256(append([]byte(password), salt[:]...))

	// generate a new private key for this user, and store it
	key := crypt.NewKey()
	err = s.keys.Set(ctx, username, keys.UniqueID, key[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create user %s: %v", username, err)
	}
	var rollback = func() error {
		return s.keys.Delete(ctx, username, keys.UniqueID)
	}

	// create the user
	u := &user.User{
		Username: username,
		Hash:     string(hashedPassword[:]),
		Salt:     string(salt[:]),
		Name:     name,
	}
	id, err := s.users.Create(ctx, u)
	if err != nil {
		rerr := rollback()
		if rerr != nil {
			err = fmt.Errorf("%w: rollback error: %v", err, rerr)
		}
		return nil, fmt.Errorf("failed to create user %s: %v", username, err)
	}

	u.ID = id
	return u, nil
}

func (s service) GetUser(ctx context.Context, username string) (*user.User, error) {
	if username == "" {
		return nil, fmt.Errorf("%w: username cannot be empty", ErrNoUser)
	}

	u, err := s.users.Get(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user %s: %v", username, err)
	}
	return u, nil
}

func (s service) ListUsers(ctx context.Context) ([]*user.User, error) {
	users, err := s.users.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %v", err)
	}
	return users, nil
}

func (s service) UpdateUser(ctx context.Context, username string, updated *user.User) error {
	if username == "" {
		return fmt.Errorf("%w: username cannot be empty", ErrNoUser)
	}
	if updated == nil {
		return fmt.Errorf("%w: updated user cannot be nil", ErrNoUser)
	}
	if updated.Name == "" {
		return fmt.Errorf("%w: new name cannot be empty", ErrNoName)
	}

	currentUser, err := s.users.Get(ctx, username)
	if err != nil {
		return fmt.Errorf("failed to fetch original user %s: %v", username, err)
	}
	if updated.Name == currentUser.Name {
		// no changes to be made
		return nil
	}

	err = s.users.Update(ctx, username, updated)
	if err != nil {
		return fmt.Errorf("failed to update user %s: %v", username, err)
	}

	return nil
}

func (s service) DeleteUser(ctx context.Context, username string) error {
	if username == "" {
		return fmt.Errorf("%w: username cannot be empty", ErrNoUser)
	}

	_, err := s.users.Get(ctx, username)
	if err != nil {
		if errors.Is(sqlite.ErrNotFoundUser, err) {
			// no change in state
			return nil
		}
		return fmt.Errorf("failed to fetch original user %s: %v", username, err)
	}

	// get user's private key for rollback func
	upk, err := s.keys.Get(ctx, username, keys.UniqueID)
	if err != nil {
		return fmt.Errorf("failed to fetch user %s's key: %v", username, err)
	}
	var rollback = func() error {
		return s.keys.Set(ctx, username, keys.UniqueID, upk)
	}

	// delete private key
	err = s.keys.Delete(ctx, username, keys.UniqueID)
	if err != nil {
		return fmt.Errorf("failed to delete user %s's key: %v", username, err)
	}

	// delete user
	err = s.users.Delete(ctx, username)
	if err != nil {
		// rollback key deletion since user deletion failed
		rerr := rollback()
		if rerr != nil {
			err = fmt.Errorf("%w: rollback error: %v", err, rerr)
		}
		return fmt.Errorf("failed to delete user %s: %v", username, err)
	}

	return nil
}
