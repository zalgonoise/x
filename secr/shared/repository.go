package shared

import (
	"context"
	"time"
)

type Repository interface {
	// Get fetches the secret's share metadata for a given username and secret key
	Get(ctx context.Context, username, secretName string) (*Shared, error)
	// Create shares the secret identified by `secretName`, owned by `owner`, with
	// user `target`. Returns an error
	Create(ctx context.Context, owner, secretName string, targets ...string) error
	// CreateUntil is similar to Create, but scopes the shared secret until `until` time
	CreateUntil(ctx context.Context, owner, secretName string, until time.Time, targets ...string) error
	// CreateFor is similar to CreateUntil, but scopes the shared secret for `dur` amount of time
	CreateFor(ctx context.Context, owner, secretName string, dur time.Duration, targets ...string) error
	// Delete removes the user `target` from the secret share
	Delete(ctx context.Context, owner, secretName string, targets ...string) error
	// Purge removes the shared secret completely so it's private to the owner again
	Purge(ctx context.Context, owner, secretName string) error
}
