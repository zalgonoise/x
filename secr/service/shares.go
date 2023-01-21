package service

import (
	"context"
	"time"

	"github.com/zalgonoise/x/secr/shared"
)

// CreateShare shares the secret with key `secretKey` belonging to user with username `owner`, with users `targets`.
// Returns the resulting shared secret, and an error
func (s service) CreateShare(ctx context.Context, owner, secretKey string, targets ...string) (*shared.Share, error) {

	return nil, nil
}

// ShareFor is similar to CreateShare, but sets the shared secret to expire after `dur` Duration
func (s service) ShareFor(ctx context.Context, owner, secretKey string, dur time.Duration, targets ...string) (*shared.Share, error) {

	return nil, nil
}

// ShareFor is similar to CreateShare, but sets the shared secret to expire after `until` Time
func (s service) ShareUntil(ctx context.Context, owner, secretKey string, until time.Time, targets ...string) (*shared.Share, error) {

	return nil, nil
}

// GetShare fetches the shared secret belonging to `username`, with key `secretKey`, returning it as a
// shared secret and an error
func (s service) GetShare(ctx context.Context, username, secretKey string) (*shared.Share, error) {

	return nil, nil
}

// DeleteShare removes the users `targets` from a shared secret with key `secretKey`, belonging to `username`. Returns
// an error
func (s service) DeleteShare(ctx context.Context, username, secretKey string, targets ...string) error {

	return nil
}

// PurgeShares removes the shared secret completely, so it's no longer available to the users it was
// shared with. Returns an error
func (s service) PurgeShares(ctx context.Context, username, secretKey string) error {

	return nil
}
