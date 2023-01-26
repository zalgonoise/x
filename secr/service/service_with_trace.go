package service

import (
	"context"
	"time"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/spanner"
	"github.com/zalgonoise/x/secr/secret"
	"github.com/zalgonoise/x/secr/shared"
	"github.com/zalgonoise/x/secr/user"
)

type withTrace struct {
	r Service
}

func WithTrace(r Service) Service {
	return withTrace{
		r: r,
	}
}

// Login verifies the user's credentials and returns a session and an error
func (t withTrace) Login(ctx context.Context, username, password string) (*user.Session, error) {
	ctx, s := spanner.Start(ctx, "service.Login")
	defer s.End()
	s.Add(
		attr.String("username", username),
	)

	sess, err := t.r.Login(ctx, username, password)
	if err != nil {
		s.Event("error logging user in", attr.New("error", err.Error()))
		return sess, err
	}
	return sess, nil
}

// Logout signs-out the user `username`
func (t withTrace) Logout(ctx context.Context, username string) error {
	ctx, s := spanner.Start(ctx, "service.Logout")
	defer s.End()
	s.Add(
		attr.String("username", username),
	)

	err := t.r.Logout(ctx, username)
	if err != nil {
		s.Event("error logging user out", attr.New("error", err.Error()))
		return err
	}
	return nil
}

// ChangePassword updates user `username`'s password after verifying the old one, returning an error
func (t withTrace) ChangePassword(ctx context.Context, username, password, newPassword string) error {
	ctx, s := spanner.Start(ctx, "service.ChangePassword")
	defer s.End()
	s.Add(
		attr.String("username", username),
	)

	err := t.r.ChangePassword(ctx, username, password, newPassword)
	if err != nil {
		s.Event("error changing user's password", attr.New("error", err.Error()))
		return err
	}
	return nil
}

// Refresh renews a user's JWT provided it is a valid one. Returns a session and an error
func (t withTrace) Refresh(ctx context.Context, username, token string) (*user.Session, error) {
	ctx, s := spanner.Start(ctx, "service.Refresh")
	defer s.End()
	s.Add(
		attr.String("username", username),
	)

	tok, err := t.r.Refresh(ctx, username, token)
	if err != nil {
		s.Event("error refreshing user's token", attr.New("error", err.Error()))
		return tok, err
	}
	return tok, nil
}

// CreateUser creates the user under username `username`, with the provided password `password` and name `name`
// It returns a user and an error
func (t withTrace) CreateUser(ctx context.Context, username, password, name string) (*user.User, error) {
	ctx, s := spanner.Start(ctx, "service.CreateUser")
	defer s.End()
	s.Add(
		attr.String("username", username),
	)

	u, err := t.r.CreateUser(ctx, username, password, name)
	if err != nil {
		s.Event("error creating user", attr.New("error", err.Error()))
		return u, err
	}
	return u, nil
}

// GetUser fetches the user with username `username`. Returns a user and an error
func (t withTrace) GetUser(ctx context.Context, username string) (*user.User, error) {
	ctx, s := spanner.Start(ctx, "service.GetUser")
	defer s.End()
	s.Add(
		attr.String("username", username),
	)

	u, err := t.r.GetUser(ctx, username)
	if err != nil {
		s.Event("error fetching user", attr.New("error", err.Error()))
		return u, err
	}
	return u, nil
}

// ListUsers returns all the users in the directory, and an error
func (t withTrace) ListUsers(ctx context.Context) ([]*user.User, error) {
	ctx, s := spanner.Start(ctx, "service.ListUsers")
	defer s.End()

	u, err := t.r.ListUsers(ctx)
	if err != nil {
		s.Event("error listing users", attr.New("error", err.Error()))
		return u, err
	}
	return u, nil
}

// UpdateUser updates the user `username`'s name, found in `updated` user. Returns an error
func (t withTrace) UpdateUser(ctx context.Context, username string, updated *user.User) error {
	ctx, s := spanner.Start(ctx, "service.GetUser")
	defer s.End()
	s.Add(
		attr.String("username", username),
		attr.String("new_name", updated.Name),
	)

	err := t.r.UpdateUser(ctx, username, updated)
	if err != nil {
		s.Event("error updating user", attr.New("error", err.Error()))
		return err
	}
	return nil
}

// DeleteUser removes the user with username `username`. Returns an error
func (t withTrace) DeleteUser(ctx context.Context, username string) error {
	ctx, s := spanner.Start(ctx, "service.DeleteUser")
	defer s.End()
	s.Add(
		attr.String("username", username),
	)

	err := t.r.DeleteUser(ctx, username)
	if err != nil {
		s.Event("error deleting user", attr.New("error", err.Error()))
		return err
	}
	return nil
}

// CreateSecret creates a secret with key `key` and value `value` (as a slice of bytes), for the
// user `username`. It returns an error
func (t withTrace) CreateSecret(ctx context.Context, username string, key string, value []byte) error {
	ctx, s := spanner.Start(ctx, "service.CreateSecret")
	defer s.End()
	s.Add(
		attr.String("username", username),
	)

	err := t.r.CreateSecret(ctx, username, key, value)
	if err != nil {
		s.Event("error creating secret", attr.New("error", err.Error()))
		return err
	}
	return nil
}

// GetSecret fetches the secret with key `key`, for user `username`. Returns a secret and an error
func (t withTrace) GetSecret(ctx context.Context, username string, key string) (*secret.Secret, error) {
	ctx, s := spanner.Start(ctx, "service.GetSecret")
	defer s.End()
	s.Add(
		attr.String("username", username),
	)

	secr, err := t.r.GetSecret(ctx, username, key)
	if err != nil {
		s.Event("error fetching secret", attr.New("error", err.Error()))
		return secr, err
	}
	return secr, nil
}

// ListSecrets retuns all secrets for user `username`. Returns a list of secrets and an error
func (t withTrace) ListSecrets(ctx context.Context, username string) ([]*secret.Secret, error) {
	ctx, s := spanner.Start(ctx, "service.ListSecrets")
	defer s.End()
	s.Add(
		attr.String("username", username),
	)

	secr, err := t.r.ListSecrets(ctx, username)
	if err != nil {
		s.Event("error listing secrets", attr.New("error", err.Error()))
		return secr, err
	}
	return secr, nil
}

// DeleteSecret removes a secret with key `key` from the user `username`. Returns an error
func (t withTrace) DeleteSecret(ctx context.Context, username string, key string) error {
	ctx, s := spanner.Start(ctx, "service.DeleteSecret")
	defer s.End()
	s.Add(
		attr.String("username", username),
	)

	err := t.r.DeleteSecret(ctx, username, key)
	if err != nil {
		s.Event("error deleting secret", attr.New("error", err.Error()))
		return err
	}
	return nil
}

// CreateShare shares the secret with key `secretKey` belonging to user with username `owner`, with users `targets`.
// Returns the resulting shared secret, and an error
func (t withTrace) CreateShare(ctx context.Context, owner, secretKey string, targets ...string) (*shared.Share, error) {
	ctx, s := spanner.Start(ctx, "service.CreateShare")
	defer s.End()
	s.Add(
		attr.String("username", owner),
	)

	id, err := t.r.CreateShare(ctx, owner, secretKey, targets...)
	if err != nil {
		s.Event("error creating shared secret", attr.New("error", err.Error()))
		return id, err
	}
	return id, nil
}

// ShareFor is similar to CreateShare, but sets the shared secret to expire after `dur` Duration
func (t withTrace) ShareFor(ctx context.Context, owner, secretKey string, dur time.Duration, targets ...string) (*shared.Share, error) {
	ctx, s := spanner.Start(ctx, "service.ShareFor")
	defer s.End()
	s.Add(
		attr.String("username", owner),
	)

	id, err := t.r.ShareFor(ctx, owner, secretKey, dur, targets...)
	if err != nil {
		s.Event("error creating shared secret with duration", attr.New("error", err.Error()))
		return id, err
	}
	return id, nil
}

// ShareFor is similar to CreateShare, but sets the shared secret to expire after `until` Time
func (t withTrace) ShareUntil(ctx context.Context, owner, secretKey string, until time.Time, targets ...string) (*shared.Share, error) {
	ctx, s := spanner.Start(ctx, "service.ShareUntil")
	defer s.End()
	s.Add(
		attr.String("username", owner),
	)

	id, err := t.r.ShareUntil(ctx, owner, secretKey, until, targets...)
	if err != nil {
		s.Event("error creating shared secret with deadline", attr.New("error", err.Error()))
		return id, err
	}
	return id, nil
}

// GetShare fetches the shared secret belonging to `username`, with key `secretKey`, returning it as a
// shared secret and an error
func (t withTrace) GetShare(ctx context.Context, username, secretKey string) ([]*shared.Share, error) {
	ctx, s := spanner.Start(ctx, "service.GetShare")
	defer s.End()
	s.Add(
		attr.String("username", username),
	)

	share, err := t.r.GetShare(ctx, username, secretKey)
	if err != nil {
		s.Event("error fetching shared secret", attr.New("error", err.Error()))
		return share, err
	}
	return share, nil
}

// ListShares fetches all the secrets the user with username `username` has shared with other users
func (t withTrace) ListShares(ctx context.Context, username string) ([]*shared.Share, error) {
	ctx, s := spanner.Start(ctx, "service.ListShares")
	defer s.End()
	s.Add(
		attr.String("username", username),
	)

	share, err := t.r.ListShares(ctx, username)
	if err != nil {
		s.Event("error listing shared secrets", attr.New("error", err.Error()))
		return share, err
	}
	return share, nil
}

// DeleteShare removes the users `targets` from a shared secret with key `secretKey`, belonging to `username`. Returns
// an error
func (t withTrace) DeleteShare(ctx context.Context, username, secretKey string, targets ...string) error {
	ctx, s := spanner.Start(ctx, "service.DeleteShare")
	defer s.End()
	s.Add(
		attr.String("username", username),
	)

	err := t.r.DeleteShare(ctx, username, secretKey, targets...)
	if err != nil {
		s.Event("error deleting shared secret", attr.New("error", err.Error()))
		return err
	}
	return nil
}

// PurgeShares removes the shared secret completely, so it's no longer available to the users it was
// shared with. Returns an error
func (t withTrace) PurgeShares(ctx context.Context, username, secretKey string) error {
	ctx, s := spanner.Start(ctx, "service.PurgeShares")
	defer s.End()
	s.Add(
		attr.String("username", username),
	)

	err := t.r.PurgeShares(ctx, username, secretKey)
	if err != nil {
		s.Event("error purging shared secret", attr.New("error", err.Error()))
		return err
	}
	return nil
}
