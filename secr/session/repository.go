package session

import (
	"context"

	"github.com/zalgonoise/x/secr/user"
)

type Repository interface {
	Login(ctx context.Context, u *user.User, password string) (*Session, error)
	Logout(ctx context.Context, u *user.User) error
	ChangePassword(ctx context.Context, u *user.User, password, newPassword string) error
	Refresh(ctx context.Context, u *user.User, token string) (*Session, error)
}
