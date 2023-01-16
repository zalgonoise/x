package secret

import (
	"context"
)

// Repository describes the actions exposed by the secrets store
type Repository interface {
	// Create will create (or overwrite) the secret identified by `s.Key`, for user `username`,
	// returning an error
	Create(ctx context.Context, username string, s *Secret) error
	// Get fetches a secret identified by `key` for user `username`. Returns a secret and an error
	Get(ctx context.Context, username string, key string) (*Secret, error)
	// List returns all secrets belonging to user `username`, and an error
	List(ctx context.Context, username string) ([]*Secret, error)
	// Delete removes the secret identified by `key`, for user `username`. Returns an error
	Delete(ctx context.Context, username string, key string) error

	// // Share shares the secret identified by `secretName`, owned by `owner`, with
	// // user `target`. Returns an error
	// Share(ctx context.Context, owner, target, secretName string) error
	// // ShareUntil is similar to Share, but scopes the shared secret until `until` time
	// ShareUntil(ctx context.Context, owner, target, secretName string, until time.Time) error
	// // ShareFor is similar to ShareUntil, but scopes the shared secret for `dur` amount of time
	// ShareUntil(ctx context.Context, owner, target, secretName string, dur time.Duration) error
}
