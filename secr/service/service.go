package service

import (
	"context"

	"github.com/zalgonoise/x/secr/secret"
	"github.com/zalgonoise/x/secr/session"
	"github.com/zalgonoise/x/secr/user"
)

type Service interface {
	// user accounts
	SignUp(ctx context.Context, username, password, name string) (*user.User, error)
	ChangePassword(ctx context.Context, username, password, newPassword string) error
	Delete(ctx context.Context, username string) error

	// user directory
	GetUser(ctx context.Context, username string) (*user.User, error)
	ListUsers(ctx context.Context) ([]*user.User, error)
	UpdateUser(ctx context.Context, username string, updated *user.User) error

	// user sessions
	Login(ctx context.Context, username, password string) (*session.Session, error)
	Logout(ctx context.Context, username string) error
	Refresh(ctx context.Context, username, token string) (*session.Session, error)

	// secrets
	CreateSecret(ctx context.Context, username string, key string, value []byte) error
	GetSecret(ctx context.Context, username string, key string) (*secret.Secret, error)
	ListSecrets(ctx context.Context, username string) ([]*secret.Secret, error)
	DeleteSecret(ctx context.Context, username string, key string) error
}
