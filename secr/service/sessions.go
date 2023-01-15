package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/zalgonoise/x/secr/authz"
	"github.com/zalgonoise/x/secr/keys"
	"github.com/zalgonoise/x/secr/user"
)

var (
	ErrIncorrectPassword = errors.New("invalid username or password")
)

func (s service) login(ctx context.Context, u *user.User, password string) error {
	// check password against its hash or as JWT
	hashedPassword := sha256.Sum256(append([]byte(password), []byte(u.Salt)...))
	if string(hashedPassword[:]) != u.Hash {
		ok, err := s.auth.Validate(ctx, u, password)
		if err != nil {
			if errors.Is(authz.ErrExpired, err) {
				derr := s.keys.Delete(ctx, u.Username, keys.TokenKey)
				if derr != nil {
					err = fmt.Errorf("%w: failed to remove old token: %v", err, derr)
				}
			}
			return fmt.Errorf("failed to validate JWT: %v", err)
		}
		if !ok {
			return fmt.Errorf("%w: couldn't login with username %s", ErrIncorrectPassword, u.Username)
		}
	}
	return nil
}

// Login verifies the user's credentials and returns a session and an error
func (s service) Login(ctx context.Context, username, password string) (*user.Session, error) {
	if username == "" {
		return nil, fmt.Errorf("%w: username cannot be empty", ErrNoUser)
	}
	if password == "" {
		return nil, fmt.Errorf("%w: password cannot be empty", ErrNoPassword)
	}

	// fetch user
	u, err := s.users.Get(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user %s: %v", username, err)
	}

	// validate credentials
	if err := s.login(ctx, u, password); err != nil {
		return nil, fmt.Errorf("%w: failed to validate user credentials: %v", ErrIncorrectPassword, err)
	}

	// issue token
	token, err := s.auth.NewToken(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new session token: %v", err)
	}

	err = s.keys.Set(ctx, u.Username, keys.TokenKey, []byte(token))
	if err != nil {
		return nil, fmt.Errorf("failed to store the new session token: %v", err)
	}

	return &user.Session{
		User:  *u,
		Token: token,
	}, nil
}

// Logout signs-out the user `username`
func (s service) Logout(ctx context.Context, username string) error {
	err := s.keys.Delete(ctx, username, keys.TokenKey)
	if err != nil {
		return fmt.Errorf("failed to log user out: %v", err)
	}
	return nil
}

// ChangePassword updates user `username`'s password after verifying the old one, returning an error
func (s service) ChangePassword(ctx context.Context, username, password, newPassword string) error {
	if username == "" {
		return fmt.Errorf("%w: username cannot be empty", ErrNoUser)
	}
	if password == "" {
		return fmt.Errorf("%w: password cannot be empty", ErrNoPassword)
	}
	if newPassword == "" {
		return fmt.Errorf("%w: new password cannot be empty", ErrNoPassword)
	}

	// fetch user
	u, err := s.users.Get(ctx, username)
	if err != nil {
		return fmt.Errorf("failed to fetch user %s: %v", username, err)
	}

	if err := s.login(ctx, u, password); err != nil {
		return fmt.Errorf("%w: failed to validate user credentials: %v", ErrIncorrectPassword, err)
	}

	hashedPassword := sha256.Sum256(append([]byte(newPassword), []byte(u.Salt)...))
	u.Hash = string(hashedPassword[:])

	err = s.users.Update(ctx, username, u)
	if err != nil {
		return fmt.Errorf("failed to update user %s's password: %v", username, err)
	}
	return nil
}

// Refresh renews a user's JWT provided it is a valid one. Returns a session and an error
func (s service) Refresh(ctx context.Context, username, token string) (*user.Session, error) {
	if username == "" {
		return nil, fmt.Errorf("%w: username cannot be empty", ErrNoUser)
	}
	if token == "" {
		return nil, fmt.Errorf("%w: token cannot be empty", ErrNoPassword)
	}

	// fetch user
	u, err := s.users.Get(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user %s: %v", username, err)
	}

	newToken, err := s.auth.Refresh(ctx, u, token)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %v", err)
	}

	err = s.keys.Set(ctx, u.Username, keys.TokenKey, []byte(newToken))
	if err != nil {
		return nil, fmt.Errorf("failed to store the new session token: %v", err)
	}

	return &user.Session{
		User:  *u,
		Token: newToken,
	}, nil
}

// Validate verifies if a user's JWT is a valid one, returning a boolean and an error
func (s service) Validate(ctx context.Context, username, token string) (bool, error) {
	if username == "" {
		return false, fmt.Errorf("%w: username cannot be empty", ErrNoUser)
	}
	if token == "" {
		return false, fmt.Errorf("%w: password cannot be empty", ErrNoPassword)
	}

	// fetch user
	u, err := s.users.Get(ctx, username)
	if err != nil {
		return false, fmt.Errorf("failed to fetch user %s: %v", username, err)
	}

	// validate credentials
	if err := s.login(ctx, u, token); err != nil {
		return false, fmt.Errorf("%w: failed to validate user credentials: %v", ErrIncorrectPassword, err)
	}
	return true, nil
}

func (s service) ParseToken(ctx context.Context, token string) (*user.User, error) {
	if token == "" {
		return nil, fmt.Errorf("%w: password cannot be empty", ErrNoPassword)
	}

	t, err := s.auth.Parse(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	u, err := s.users.Get(ctx, t.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user %s: %v", t.Username, err)
	}

	return u, nil
}
