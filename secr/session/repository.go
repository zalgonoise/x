package session

import (
	"context"
)

type Repository interface {
	Login(ctx context.Context, username, password string) (*Session, error)
	Logout(ctx context.Context, username string) error
	ChangePassword(ctx context.Context, username, password, newPassword string) error
	Refresh(ctx context.Context, username, token string) (*Session, error)
}
