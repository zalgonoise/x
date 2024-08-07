package service

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/zalgonoise/x/errors"
	"github.com/zalgonoise/x/secr/crypt"
	"github.com/zalgonoise/x/secr/keys"
	"github.com/zalgonoise/x/secr/user"
)

var (
	ErrIncorrectPassword = errors.New("invalid username or password")
)

func (s service) login(ctx context.Context, u *user.User, password string) error {
	hash, err := base64.StdEncoding.DecodeString(u.Hash)
	if err != nil {
		return fmt.Errorf("failed to decode hash: %w", err)
	}
	salt, err := base64.StdEncoding.DecodeString(u.Salt)
	if err != nil {
		return fmt.Errorf("failed to decode salt: %w", err)
	}
	hashedPassword := crypt.Hash([]byte(password), salt)

	if string(hashedPassword[:]) != string(hash) {
		return ErrIncorrectPassword
	}
	return nil
}

// Login verifies the user's credentials and returns a session and an error
func (s service) Login(ctx context.Context, username, password string) (*user.Session, error) {
	if err := user.ValidateUsername(username); err != nil {
		return nil, errors.Join(ErrInvalidUser, err)
	}
	if err := user.ValidatePassword(password); err != nil {
		return nil, errors.Join(ErrInvalidPassword, err)
	}

	// fetch user
	u, err := s.users.Get(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user %s: %w", username, err)
	}

	// validate credentials
	if err := s.login(ctx, u, password); err != nil {
		return nil, fmt.Errorf("failed to validate user credentials: %w", err)
	}

	// issue token
	token, err := s.auth.NewToken(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new session token: %w", err)
	}

	err = s.keys.Set(ctx, keys.UserBucket(u.ID), keys.TokenKey, []byte(token))
	if err != nil {
		return nil, fmt.Errorf("failed to store the new session token: %w", err)
	}

	return &user.Session{
		User:  *u,
		Token: token,
	}, nil
}

// Logout signs-out the user `username`
func (s service) Logout(ctx context.Context, username string) error {
	if err := user.ValidateUsername(username); err != nil {
		return errors.Join(ErrInvalidUser, err)
	}

	u, err := s.users.Get(ctx, username)
	if err != nil {
		return fmt.Errorf("failed to fetch user: %w", err)
	}

	err = s.keys.Delete(ctx, keys.UserBucket(u.ID), keys.TokenKey)
	if err != nil {
		return fmt.Errorf("failed to log user out: %w", err)
	}
	return nil
}

// ChangePassword updates user `username`'s password after verifying the old one, returning an error
func (s service) ChangePassword(ctx context.Context, username, password, newPassword string) error {
	if err := user.ValidateUsername(username); err != nil {
		return errors.Join(ErrInvalidUser, err)
	}
	if err := user.ValidatePassword(password); err != nil {
		return errors.Join(ErrInvalidPassword, err)
	}
	if err := user.ValidatePassword(newPassword); err != nil {
		return errors.Join(ErrInvalidPassword, err)
	}

	// fetch user
	u, err := s.users.Get(ctx, username)
	if err != nil {
		return fmt.Errorf("failed to fetch user %s: %w", username, err)
	}

	if err := s.login(ctx, u, password); err != nil {
		return fmt.Errorf("failed to validate user credentials: %w", err)
	}

	salt, err := base64.StdEncoding.DecodeString(u.Salt)
	if err != nil {
		return fmt.Errorf("failed to decode salt: %w", err)
	}

	u.Hash = string(crypt.Hash([]byte(password), salt))
	err = s.users.Update(ctx, username, u)
	if err != nil {
		return fmt.Errorf("failed to update user %s's password: %w", username, err)
	}
	return nil
}

// Refresh renews a user's JWT provided it is a valid one. Returns a session and an error
func (s service) Refresh(ctx context.Context, username, token string) (*user.Session, error) {
	if err := user.ValidateUsername(username); err != nil {
		return nil, errors.Join(ErrInvalidUser, err)
	}
	if token == "" {
		return nil, fmt.Errorf("%w: token cannot be empty", ErrInvalidPassword)
	}

	// fetch user
	u, err := s.users.Get(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user %s: %w", username, err)
	}

	jwtUser, err := s.auth.Parse(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	if jwtUser.Username != u.Username {
		return nil, ErrIncorrectPassword
	}

	newToken, err := s.auth.NewToken(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	err = s.keys.Set(ctx, keys.UserBucket(u.ID), keys.TokenKey, []byte(newToken))
	if err != nil {
		return nil, fmt.Errorf("failed to store the new session token: %w", err)
	}

	return &user.Session{
		User:  *u,
		Token: newToken,
	}, nil
}

// ParseToken reads the input token string and returns the corresponding user in it, or an error
func (s service) ParseToken(ctx context.Context, token string) (*user.User, error) {
	if token == "" {
		return nil, fmt.Errorf("%w: token cannot be empty", ErrInvalidPassword)
	}

	t, err := s.auth.Parse(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	u, err := s.users.Get(ctx, t.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user %s: %w", t.Username, err)
	}

	return u, nil
}
