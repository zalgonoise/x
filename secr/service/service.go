package service

import (
	"context"

	"github.com/zalgonoise/x/secr/keys"
	"github.com/zalgonoise/x/secr/secret"
	"github.com/zalgonoise/x/secr/session"
	"github.com/zalgonoise/x/secr/user"
)

type Service interface {
	// user sessions
	Login(ctx context.Context, username, password string) (*session.Session, error)
	Logout(ctx context.Context, username string) error
	ChangePassword(ctx context.Context, username, password, newPassword string) error
	Refresh(ctx context.Context, username, token string) (*session.Session, error)

	// user directory
	CreateUser(ctx context.Context, username, password, name string) (*user.User, error)
	GetUser(ctx context.Context, username string) (*user.User, error)
	ListUsers(ctx context.Context) ([]*user.User, error)
	UpdateUser(ctx context.Context, username string, updated *user.User) error
	DeleteUser(ctx context.Context, username string) error

	// secrets
	CreateSecret(ctx context.Context, username string, key string, value []byte) error
	GetSecret(ctx context.Context, username string, key string) (*secret.Secret, error)
	ListSecrets(ctx context.Context, username string) ([]*secret.Secret, error)
	DeleteSecret(ctx context.Context, username string, key string) error
}

type service struct {
	users    user.Repository
	secrets  secret.Repository
	keys     keys.Repository
	sessions session.Repository
}

func NewService(
	u user.Repository,
	s secret.Repository,
	k keys.Repository,
	ss session.Repository,
) Service {
	return service{
		users:    u,
		secrets:  s,
		keys:     k,
		sessions: ss,
	}
}
