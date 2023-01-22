package service

import (
	"context"
	"time"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/logx"
	"github.com/zalgonoise/x/secr/secret"
	"github.com/zalgonoise/x/secr/shared"
	"github.com/zalgonoise/x/secr/user"
)

type withLogger struct {
	r Service
	l logx.Logger
}

func WithLogger(l logx.Logger, r Service) Service {
	return withLogger{
		r: r,
		l: l,
	}
}

// Login verifies the user's credentials and returns a session and an error
func (l withLogger) Login(ctx context.Context, username, password string) (*user.Session, error) {
	sess, err := l.r.Login(ctx, username, password)
	if err != nil {
		l.l.Error(
			err.Error(),
			attr.String("service", "service.Login"),
			attr.String("username", username),
		)
		return sess, err
	}
	return sess, nil
}

// Logout signs-out the user `username`
func (l withLogger) Logout(ctx context.Context, username string) error {
	err := l.r.Logout(ctx, username)
	if err != nil {
		l.l.Error(
			err.Error(),
			attr.String("service", "service.Logout"),
			attr.String("username", username),
		)
		return err
	}
	return nil
}

// ChangePassword updates user `username`'s password after verifying the old one, returning an error
func (l withLogger) ChangePassword(ctx context.Context, username, password, newPassword string) error {
	err := l.r.ChangePassword(ctx, username, password, newPassword)
	if err != nil {
		l.l.Error(
			err.Error(),
			attr.String("service", "service.ChangePassword"),
			attr.String("username", username),
		)
		return err
	}
	return nil
}

// Refresh renews a user's JWT provided it is a valid one. Returns a session and an error
func (l withLogger) Refresh(ctx context.Context, username, token string) (*user.Session, error) {
	tok, err := l.r.Refresh(ctx, username, token)
	if err != nil {
		l.l.Error(
			err.Error(),
			attr.String("service", "service.Refresh"),
			attr.String("username", username),
		)
		return tok, err
	}
	return tok, nil
}

// Validate verifies if a user's JWT is a valid one, returning a boolean and an error
func (l withLogger) Validate(ctx context.Context, username, token string) (bool, error) {
	ok, err := l.r.Validate(ctx, username, token)
	if err != nil {
		l.l.Error(
			err.Error(),
			attr.String("service", "service.Validate"),
			attr.String("username", username),
		)
		return ok, err
	}
	return ok, nil
}

func (l withLogger) ParseToken(ctx context.Context, token string) (*user.User, error) {
	u, err := l.r.ParseToken(ctx, token)
	if err != nil {
		l.l.Error(
			err.Error(),
			attr.String("service", "service.ParseToken"),
		)
		return u, err
	}
	return u, nil
}

// CreateUser creates the user under username `username`, with the provided password `password` and name `name`
// It returns a user and an error
func (l withLogger) CreateUser(ctx context.Context, username, password, name string) (*user.User, error) {
	u, err := l.r.CreateUser(ctx, username, password, name)
	if err != nil {
		l.l.Error(
			err.Error(),
			attr.String("service", "service.CreateUser"),
			attr.String("username", username),
		)
		return u, err
	}
	return u, nil
}

// GetUser fetches the user with username `username`. Returns a user and an error
func (l withLogger) GetUser(ctx context.Context, username string) (*user.User, error) {
	u, err := l.r.GetUser(ctx, username)
	if err != nil {
		l.l.Error(
			err.Error(),
			attr.String("service", "service.GetUser"),
			attr.String("username", username),
		)
		return u, err
	}
	return u, nil
}

// ListUsers returns all the users in the directory, and an error
func (l withLogger) ListUsers(ctx context.Context) ([]*user.User, error) {
	u, err := l.r.ListUsers(ctx)
	if err != nil {
		l.l.Error(
			err.Error(),
			attr.String("service", "service.ListUsers"),
		)
		return u, err
	}
	return u, nil
}

// UpdateUser updates the user `username`'s name, found in `updated` user. Returns an error
func (l withLogger) UpdateUser(ctx context.Context, username string, updated *user.User) error {
	err := l.r.UpdateUser(ctx, username, updated)
	if err != nil {
		l.l.Error(
			err.Error(),
			attr.String("service", "service.UpdateUser"),
			attr.String("username", username),
		)
		return err
	}
	return nil
}

// DeleteUser removes the user with username `username`. Returns an error
func (l withLogger) DeleteUser(ctx context.Context, username string) error {
	err := l.r.DeleteUser(ctx, username)
	if err != nil {
		l.l.Error(
			err.Error(),
			attr.String("service", "service.DeleteUser"),
			attr.String("username", username),
		)
		return err
	}
	return nil
}

// CreateSecret creates a secret with key `key` and value `value` (as a slice of bytes), for the
// user `username`. It returns an error
func (l withLogger) CreateSecret(ctx context.Context, username string, key string, value []byte) error {
	err := l.r.CreateSecret(ctx, username, key, value)
	if err != nil {
		l.l.Error(
			err.Error(),
			attr.String("service", "service.CreateSecret"),
			attr.String("username", username),
		)
		return err
	}
	return nil
}

// GetSecret fetches the secret with key `key`, for user `username`. Returns a secret and an error
func (l withLogger) GetSecret(ctx context.Context, username string, key string) (*secret.Secret, error) {
	secr, err := l.r.GetSecret(ctx, username, key)
	if err != nil {
		l.l.Error(
			err.Error(),
			attr.String("service", "service.GetSecret"),
			attr.String("username", username),
		)
		return secr, err
	}
	return secr, nil
}

// ListSecrets retuns all secrets for user `username`. Returns a list of secrets and an error
func (l withLogger) ListSecrets(ctx context.Context, username string) ([]*secret.Secret, error) {
	secr, err := l.r.ListSecrets(ctx, username)
	if err != nil {
		l.l.Error(
			err.Error(),
			attr.String("service", "service.ListSecrets"),
			attr.String("username", username),
		)
		return secr, err
	}
	return secr, nil
}

// DeleteSecret removes a secret with key `key` from the user `username`. Returns an error
func (l withLogger) DeleteSecret(ctx context.Context, username string, key string) error {
	err := l.r.DeleteSecret(ctx, username, key)
	if err != nil {
		l.l.Error(
			err.Error(),
			attr.String("service", "service.DeleteSecret"),
			attr.String("username", username),
		)
		return err
	}
	return nil
}

// CreateShare shares the secret with key `secretKey` belonging to user with username `owner`, with users `targets`.
// Returns the resulting shared secret, and an error
func (l withLogger) CreateShare(ctx context.Context, owner, secretKey string, targets ...string) (*shared.Share, error) {
	share, err := l.r.CreateShare(ctx, owner, secretKey, targets...)
	if err != nil {
		l.l.Error(
			err.Error(),
			attr.String("service", "service.CreateShare"),
			attr.String("username", owner),
			attr.String("secretKey", secretKey),
			attr.New[[]string]("targets", targets),
		)
		return share, err
	}
	return share, nil
}

// ShareFor is similar to CreateShare, but sets the shared secret to expire after `dur` Duration
func (l withLogger) ShareFor(ctx context.Context, owner, secretKey string, dur time.Duration, targets ...string) (*shared.Share, error) {
	share, err := l.r.ShareFor(ctx, owner, secretKey, dur, targets...)
	if err != nil {
		l.l.Error(
			err.Error(),
			attr.String("service", "service.ShareFor"),
			attr.String("username", owner),
			attr.String("secretKey", secretKey),
			attr.String("duration", dur.String()),
			attr.New[[]string]("targets", targets),
		)
		return share, err
	}
	return share, nil
}

// ShareFor is similar to CreateShare, but sets the shared secret to expire after `until` Time
func (l withLogger) ShareUntil(ctx context.Context, owner, secretKey string, until time.Time, targets ...string) (*shared.Share, error) {
	share, err := l.r.ShareUntil(ctx, owner, secretKey, until, targets...)
	if err != nil {
		l.l.Error(
			err.Error(),
			attr.String("service", "service.ShareUntil"),
			attr.String("username", owner),
			attr.String("secretKey", secretKey),
			attr.String("deadline", until.String()),
			attr.New[[]string]("targets", targets),
		)
		return share, err
	}
	return share, nil
}

// GetShare fetches the shared secret belonging to `username`, with key `secretKey`, returning it as a
// shared secret and an error
func (l withLogger) GetShare(ctx context.Context, username, secretKey string) ([]*shared.Share, error) {
	share, err := l.r.GetShare(ctx, username, secretKey)
	if err != nil {
		l.l.Error(
			err.Error(),
			attr.String("service", "service.GetShare"),
			attr.String("username", username),
			attr.String("secretKey", secretKey),
		)
		return share, err
	}
	return share, nil
}

// ListShares fetches all the secrets the user with username `username` has shared with other users
func (l withLogger) ListShares(ctx context.Context, username string) ([]*shared.Share, error) {
	share, err := l.r.ListShares(ctx, username)
	if err != nil {
		l.l.Error(
			err.Error(),
			attr.String("service", "service.ListShares"),
			attr.String("username", username),
		)
		return share, err
	}
	return share, nil
}

// DeleteShare removes the users `targets` from a shared secret with key `secretKey`, belonging to `username`. Returns
// an error
func (l withLogger) DeleteShare(ctx context.Context, username, secretKey string, targets ...string) error {
	err := l.r.DeleteShare(ctx, username, secretKey, targets...)
	if err != nil {
		l.l.Error(
			err.Error(),
			attr.String("service", "service.DeleteShare"),
			attr.String("username", username),
			attr.String("secretKey", secretKey),
			attr.New[[]string]("targets", targets),
		)
		return err
	}
	return nil
}

// PurgeShares removes the shared secret completely, so it's no longer available to the users it was
// shared with. Returns an error
func (l withLogger) PurgeShares(ctx context.Context, username, secretKey string) error {
	err := l.r.PurgeShares(ctx, username, secretKey)
	if err != nil {
		l.l.Error(
			err.Error(),
			attr.String("service", "service.PurgeShares"),
			attr.String("username", username),
			attr.String("secretKey", secretKey),
		)
		return err
	}
	return nil
}
