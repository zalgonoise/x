package service

import (
	"context"

	"github.com/zalgonoise/x/secr/user"
)

func (s service) CreateUser(ctx context.Context, username, password, name string) (*user.User, error)
func (s service) GetUser(ctx context.Context, username string) (*user.User, error)
func (s service) ListUsers(ctx context.Context) ([]*user.User, error)
func (s service) UpdateUser(ctx context.Context, username string, updated *user.User) error
func (s service) DeleteUser(ctx context.Context, username string) error
