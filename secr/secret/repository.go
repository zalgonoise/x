package secret

import "context"

type Repository interface {
	Get(ctx context.Context, username string, key string) (*Secret, error)
	List(ctx context.Context, username string) ([]*Secret, error)
	Create(ctx context.Context, username string, key string, value []byte) error
	Delete(ctx context.Context, username string, key string) error
}
