package secret

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, username string, s *Secret) error
	Get(ctx context.Context, username string, key string) (*Secret, error)
	List(ctx context.Context, username string) ([]*Secret, error)
	Delete(ctx context.Context, username string, key string) error

	// Share(ctx context.Context, owner, target, secretName string) error
	// ShareFor(ctx context.Context, owner, target, secretName string, until time.Time) error
}
