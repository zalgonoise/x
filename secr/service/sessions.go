package service

import (
	"context"

	"github.com/zalgonoise/x/secr/session"
)

func (s service) Login(ctx context.Context, username, password string) (*session.Session, error)
func (s service) Logout(ctx context.Context, username string) error
func (s service) ChangePassword(ctx context.Context, username, password, newPassword string) error
func (s service) Refresh(ctx context.Context, username, token string) (*session.Session, error)
