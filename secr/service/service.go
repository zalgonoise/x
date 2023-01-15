package service

import (
	"context"

	"github.com/zalgonoise/x/secr/authz"
	"github.com/zalgonoise/x/secr/keys"
	"github.com/zalgonoise/x/secr/secret"
	"github.com/zalgonoise/x/secr/user"
)

// Service defines all the exposed features and functionalities of the secrets store
type Service interface {
	// Login verifies the user's credentials and returns a session and an error
	Login(ctx context.Context, username, password string) (*user.Session, error)
	// Logout signs-out the user `username`
	Logout(ctx context.Context, username string) error
	// ChangePassword updates user `username`'s password after verifying the old one, returning an error
	ChangePassword(ctx context.Context, username, password, newPassword string) error
	// Refresh renews a user's JWT provided it is a valid one. Returns a session and an error
	Refresh(ctx context.Context, username, token string) (*user.Session, error)
	// Validate verifies if a user's JWT is a valid one, returning a boolean and an error
	Validate(ctx context.Context, username, token string) (bool, error)

	// CreateUser creates the user under username `username`, with the provided password `password` and name `name`
	// It returns a user and an error
	CreateUser(ctx context.Context, username, password, name string) (*user.User, error)
	// GetUser fetches the user with username `username`. Returns a user and an error
	GetUser(ctx context.Context, username string) (*user.User, error)
	// ListUsers returns all the users in the directory, and an error
	ListUsers(ctx context.Context) ([]*user.User, error)
	// UpdateUser updates the user `username`'s name, found in `updated` user. Returns an error
	UpdateUser(ctx context.Context, username string, updated *user.User) error
	// DeleteUser removes the user with username `username`. Returns an error
	DeleteUser(ctx context.Context, username string) error

	// CreateSecret creates a secret with key `key` and value `value` (as a slice of bytes), for the
	// user `username`. It returns an error
	CreateSecret(ctx context.Context, username string, key string, value []byte) error
	// GetSecret fetches the secret with key `key`, for user `username`. Returns a secret and an error
	GetSecret(ctx context.Context, username string, key string) (*secret.Secret, error)
	// ListSecrets retuns all secrets for user `username`. Returns a list of secrets and an error
	ListSecrets(ctx context.Context, username string) ([]*secret.Secret, error)
	// DeleteSecret removes a secret with key `key` from the user `username`. Returns an error
	DeleteSecret(ctx context.Context, username string, key string) error
}

type service struct {
	users   user.Repository
	secrets secret.Repository
	keys    keys.Repository
	auth    authz.Authorizer
}

func NewService(
	users user.Repository,
	secrets secret.Repository,
	keys keys.Repository,
	auth authz.Authorizer,

) Service {
	return service{
		users:   users,
		secrets: secrets,
		keys:    keys,
		auth:    auth,
	}
}
