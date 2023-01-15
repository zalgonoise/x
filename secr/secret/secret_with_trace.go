package secret

import (
	"context"

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

// Create will create (or overwrite) the secret identified by `s.Key`, for user `username`,
// returning an error
func (t withTrace) Create(ctx context.Context, username string, secr *Secret) error {
	ctx, s := spanner.Start(ctx, "secret.Create")
	defer s.End()
	s.Add(
		attr.String("for_user", username),
	)

	err := t.r.Create(ctx, username, secr)
	if err != nil {
		s.Event("error creating secret", attr.New("error", err.Error()))
		return err
	}
	return nil
}

// Get fetches a secret identified by `key` for user `username`. Returns a secret and an error
func (t withTrace) Get(ctx context.Context, username string, key string) (*Secret, error) {
	ctx, s := spanner.Start(ctx, "secret.Create")
	defer s.End()
	s.Add(
		attr.String("for_user", username),
	)

	secr, err := t.r.Get(ctx, username, key)
	if err != nil {
		s.Event("error fetching secret", attr.New("error", err.Error()))
		return nil, err
	}
	return secr, nil
}

// List returns all secrets belonging to user `username`, and an error
func (t withTrace) List(ctx context.Context, username string) ([]*Secret, error) {
	ctx, s := spanner.Start(ctx, "secret.Create")
	defer s.End()
	s.Add(
		attr.String("for_user", username),
	)

	secr, err := t.r.List(ctx, username)
	if err != nil {
		s.Event("error listing secrets", attr.New("error", err.Error()))
		return nil, err
	}
	return secr, nil
}

// Delete removes the secret identified by `key`, for user `username`. Returns an error
func (t withTrace) Delete(ctx context.Context, username string, key string) error {
	ctx, s := spanner.Start(ctx, "secret.Create")
	defer s.End()
	s.Add(
		attr.String("for_user", username),
	)

	err := t.r.Delete(ctx, username, key)
	if err != nil {
		s.Event("error listing secrets", attr.New("error", err.Error()))
		return err
	}
	return nil
}
