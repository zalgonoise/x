package keys

import "context"

type Repository interface {
	Get(ctx context.Context, username string) (Key, error)
	Set(ctx context.Context, username string, k Key) error
	Delete(ctx context.Context, username string) error
}
