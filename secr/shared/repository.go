package shared

import (
	"context"
)

// Repository describes the actions exposed by the shared secrets store
type Repository interface {
	// Get fetches the secret's share metadata for a given owner's username and secret key
	Get(ctx context.Context, owner, secretName string) ([]*Share, error)
	// List fetches all shared secrets for a given owner's username
	List(ctx context.Context, owner string) ([]*Share, error)
	// ListTarget is similar to List, but returns secrets that are shared with a target user
	ListTarget(ctx context.Context, target string) ([]*Share, error)
	// Create shares the secret identified by `secretName`, owned by `owner`, with
	// user `target`. Returns an error
	Create(ctx context.Context, s *Share) error
	// Delete removes the user `target` from the secret share
	Delete(ctx context.Context, s *Share) error
}
