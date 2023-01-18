package shared

import (
	"context"
)

type Repository interface {
	// Get fetches the secret's share metadata for a given username and secret key
	Get(ctx context.Context, username, secretName string) ([]*Share, error)
	// List fetches all shared secrets for a given username
	List(ctx context.Context, username string) ([]*Share, error)
	// Create shares the secret identified by `secretName`, owned by `owner`, with
	// user `target`. Returns an error
	Create(ctx context.Context, s *Share) error
	// Delete removes the user `target` from the secret share
	Delete(ctx context.Context, s *Share) error
}
