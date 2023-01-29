package service

import (
	"context"
	"fmt"
	"time"

	"github.com/zalgonoise/x/errors"
	"github.com/zalgonoise/x/ptr"
	"github.com/zalgonoise/x/secr/secret"
	"github.com/zalgonoise/x/secr/shared"
	"github.com/zalgonoise/x/secr/sqlite"
	"github.com/zalgonoise/x/secr/user"
)

var (
	ErrZeroTargets = errors.New("shared secret must have at least a target user to share the secret with")
	ErrInvalidTime = errors.New("invalid shared secret time limit")
)

// CreateShare shares the secret with key `secretKey` belonging to user with username `owner`, with users `targets`.
// Returns the resulting shared secret, and an error
func (s service) CreateShare(ctx context.Context, owner, secretKey string, targets ...string) (*shared.Share, error) {
	if err := user.ValidateUsername(owner); err != nil {
		return nil, errors.Join(ErrInvalidUser, err)
	}
	if isShared, err := secret.ValidateKey(secretKey); err != nil || isShared {
		return nil, errors.Join(ErrInvalidKey, err)
	}

	if len(targets) == 0 {
		return nil, ErrZeroTargets
	}
	sh := &shared.Share{
		SecretKey: secretKey,
		Owner:     owner,
		Until:     ptr.To(time.Now().Add(shared.DefaultShareDuration)),
	}
	for _, t := range targets {
		if err := user.ValidateUsername(t); err != nil {
			return nil, errors.Join(ErrInvalidUser, err)
		}
		sh.Target = append(sh.Target, t)
	}
	err := s.shares.Delete(ctx, sh)
	if err != nil && !errors.Is(sqlite.ErrNotFoundShare, err) {
		return nil, fmt.Errorf("failed to remove previous shared secrets: %w", err)
	}

	id, err := s.shares.Create(ctx, sh)
	if err != nil {
		return nil, fmt.Errorf("failed to create shared secret: %w", err)
	}

	sh.ID = id
	return sh, nil
}

// ShareFor is similar to CreateShare, but sets the shared secret to expire after `dur` Duration
func (s service) ShareFor(ctx context.Context, owner, secretKey string, dur time.Duration, targets ...string) (*shared.Share, error) {
	if err := user.ValidateUsername(owner); err != nil {
		return nil, errors.Join(ErrInvalidUser, err)
	}
	if isShared, err := secret.ValidateKey(secretKey); err != nil || isShared {
		return nil, errors.Join(ErrInvalidKey, err)
	}
	if len(targets) == 0 {
		return nil, ErrZeroTargets
	}
	if err := shared.ValidateDuration(dur); err != nil {
		return nil, errors.Join(ErrInvalidTime, err)
	}

	sh := &shared.Share{
		SecretKey: secretKey,
		Owner:     owner,
		Until:     ptr.To(time.Now().Add(dur)),
	}
	for _, t := range targets {
		if err := user.ValidateUsername(t); err != nil {
			return nil, errors.Join(ErrInvalidUser, err)
		}
		sh.Target = append(sh.Target, t)
	}

	err := s.shares.Delete(ctx, sh)
	if err != nil && !errors.Is(sqlite.ErrNotFoundShare, err) {
		return nil, fmt.Errorf("failed to remove previous shared secrets: %w", err)
	}

	id, err := s.shares.Create(ctx, sh)
	if err != nil {
		return nil, fmt.Errorf("failed to create shared secret: %w", err)
	}

	sh.ID = id
	return sh, nil
}

// ShareFor is similar to CreateShare, but sets the shared secret to expire after `until` Time
func (s service) ShareUntil(ctx context.Context, owner, secretKey string, until time.Time, targets ...string) (*shared.Share, error) {
	if err := user.ValidateUsername(owner); err != nil {
		return nil, errors.Join(ErrInvalidUser, err)
	}
	if isShared, err := secret.ValidateKey(secretKey); err != nil || isShared {
		return nil, errors.Join(ErrInvalidKey, err)
	}
	if len(targets) == 0 {
		return nil, ErrZeroTargets
	}
	if err := shared.ValidateTime(until); err != nil {
		return nil, errors.Join(ErrInvalidTime, err)
	}

	sh := &shared.Share{
		SecretKey: secretKey,
		Owner:     owner,
		Until:     &until,
	}
	for _, t := range targets {
		if err := user.ValidateUsername(t); err != nil {
			return nil, errors.Join(ErrInvalidUser, err)
		}
		sh.Target = append(sh.Target, t)
	}

	err := s.shares.Delete(ctx, sh)
	if err != nil && !errors.Is(sqlite.ErrNotFoundShare, err) {
		return nil, fmt.Errorf("failed to remove previous shared secrets: %w", err)
	}

	id, err := s.shares.Create(ctx, sh)
	if err != nil {
		return nil, fmt.Errorf("failed to create shared secret: %w", err)
	}

	sh.ID = id
	return sh, nil
}

// GetShare fetches the shared secret belonging to `username`, with key `secretKey`, returning it as a
// shared secret and an error
func (s service) GetShare(ctx context.Context, username, secretKey string) ([]*shared.Share, error) {
	if err := user.ValidateUsername(username); err != nil {
		return nil, errors.Join(ErrInvalidUser, err)
	}
	if isShared, err := secret.ValidateKey(secretKey); err != nil || isShared {
		return nil, errors.Join(ErrInvalidKey, err)
	}

	sh, err := s.shares.Get(ctx, username, secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch shared secret: %w", err)
	}
	return sh, nil
}

// ListShares fetches all the secrets the user with username `username` has shared with other users
func (s service) ListShares(ctx context.Context, username string) ([]*shared.Share, error) {
	if err := user.ValidateUsername(username); err != nil {
		return nil, errors.Join(ErrInvalidUser, err)
	}

	sh, err := s.shares.List(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to list shared secrets: %w", err)
	}
	return sh, nil
}

// DeleteShare removes the users `targets` from a shared secret with key `secretKey`, belonging to `username`. Returns
// an error
func (s service) DeleteShare(ctx context.Context, owner, secretKey string, targets ...string) error {
	if err := user.ValidateUsername(owner); err != nil {
		return errors.Join(ErrInvalidUser, err)
	}
	if isShared, err := secret.ValidateKey(secretKey); err != nil || isShared {
		return errors.Join(ErrInvalidKey, err)
	}
	if len(targets) == 0 {
		return s.PurgeShares(ctx, owner, secretKey)
	}

	sh := &shared.Share{
		SecretKey: secretKey,
		Owner:     owner,
	}
	for _, t := range targets {
		if err := user.ValidateUsername(t); err != nil {
			return errors.Join(ErrInvalidUser, err)
		}
		sh.Target = append(sh.Target, t)
	}

	err := s.shares.Delete(ctx, sh)
	if err != nil && !errors.Is(sqlite.ErrNotFoundShare, err) {
		return fmt.Errorf("failed to delete shared secret: %w", err)
	}
	return nil
}

// PurgeShares removes the shared secret completely, so it's no longer available to the users it was
// shared with. Returns an error
func (s service) PurgeShares(ctx context.Context, owner, secretKey string) error {
	if err := user.ValidateUsername(owner); err != nil {
		return errors.Join(ErrInvalidUser, err)
	}
	if isShared, err := secret.ValidateKey(secretKey); err != nil || isShared {
		return errors.Join(ErrInvalidKey, err)
	}

	sh, err := s.shares.Get(ctx, owner, secretKey)
	if err != nil {
		return fmt.Errorf("failed to fetch shared secret: %w", err)
	}

	tx := newTx()

	for _, share := range sh {
		tx.Add(func() error {
			_, err := s.shares.Create(ctx, share)
			return err
		})
		err := s.shares.Delete(ctx, share)
		if err != nil {
			return tx.Rollback(fmt.Errorf("failed to remove shared secret: %w", err))
		}
	}
	return nil
}
