package user

import "context"

type Repository interface {
	Create(ctx context.Context, u *User) (uint64, error)
	Update(ctx context.Context, username string, updated *User) error
	Get(ctx context.Context, username string) (*User, error)
	List(ctx context.Context) ([]*User, error)
	Delete(ctx context.Context, username string) error
}
