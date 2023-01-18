package shared

import (
	"context"
	"time"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/spanner"
)

type withTrace struct {
	r Repository
}

func WithTrace(r Repository) Repository {
	return withTrace{
		r: r,
	}
}

// Get fetches the secret's share metadata for a given username and secret key
func (t withTrace) Get(ctx context.Context, username, secretName string) (*Shared, error) {
	ctx, s := spanner.Start(ctx, "secret.Get")
	defer s.End()
	s.Add(
		attr.String("for_user", username),
	)

	share, err := t.r.Get(ctx, username, secretName)
	if err != nil {
		s.Event("error fetching shared secret", attr.New("error", err.Error()))
		return nil, err
	}
	return share, nil
}

// Create shares the secret identified by `secretName`, owned by `owner`, with
// user `target`. Returns an error
func (t withTrace) Create(ctx context.Context, owner, secretName string, targets ...string) error {
	ctx, s := spanner.Start(ctx, "secret.Create")
	defer s.End()
	s.Add(
		attr.String("from_user", owner),
		attr.New("to_users", targets),
	)

	err := t.r.Create(ctx, owner, secretName, targets...)
	if err != nil {
		s.Event("error creating shared secret", attr.New("error", err.Error()))
		return err
	}
	return nil
}

// CreateUntil is similar to Create, but scopes the shared secret until `until` time
func (t withTrace) CreateUntil(ctx context.Context, owner, secretName string, until time.Time, targets ...string) error {
	ctx, s := spanner.Start(ctx, "secret.CreateUntil")
	defer s.End()
	s.Add(
		attr.String("from_user", owner),
		attr.New("to_users", targets),
	)

	err := t.r.CreateUntil(ctx, owner, secretName, until, targets...)
	if err != nil {
		s.Event("error creating shared secret", attr.New("error", err.Error()))
		return err
	}
	return nil
}

// CreateFor is similar to CreateUntil, but scopes the shared secret for `dur` amount of time
func (t withTrace) CreateFor(ctx context.Context, owner, secretName string, dur time.Duration, targets ...string) error {
	ctx, s := spanner.Start(ctx, "secret.CreateFor")
	defer s.End()
	s.Add(
		attr.String("from_user", owner),
		attr.New("to_users", targets),
	)

	err := t.r.CreateFor(ctx, owner, secretName, dur, targets...)
	if err != nil {
		s.Event("error creating shared secret", attr.New("error", err.Error()))
		return err
	}
	return nil
}

// Delete removes the user `target` from the secret share
func (t withTrace) Delete(ctx context.Context, owner, secretName string, targets ...string) error {
	ctx, s := spanner.Start(ctx, "secret.Delete")
	defer s.End()
	s.Add(
		attr.String("from_user", owner),
	)

	err := t.r.Delete(ctx, owner, secretName, targets...)
	if err != nil {
		s.Event("error deleting shared secret", attr.New("error", err.Error()))
		return err
	}
	return nil
}

// Purge removes the shared secret completely so it's private to the owner again
func (t withTrace) Purge(ctx context.Context, owner, secretName string) error {
	ctx, s := spanner.Start(ctx, "secret.Purge")
	defer s.End()
	s.Add(
		attr.String("from_user", owner),
	)

	err := t.r.Purge(ctx, owner, secretName)
	if err != nil {
		s.Event("error purging shared secrets", attr.New("error", err.Error()))
		return err
	}
	return nil
}
