package shared

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

// Create shares the secret identified by `secretName`, owned by `owner`, with
// user `target`. Returns an error
func (t withTrace) Create(ctx context.Context, sh *Share) (uint64, error) {
	ctx, s := spanner.Start(ctx, "secret.Create")
	defer s.End()
	s.Add(
		attr.String("from_user", sh.Owner.Username),
	)

	id, err := t.r.Create(ctx, sh)
	if err != nil {
		s.Event("error creating shared secret", attr.New("error", err.Error()))
		return id, err
	}
	return id, nil
}

// Get fetches the secret's share metadata for a given username and secret key
func (t withTrace) Get(ctx context.Context, username, secretName string) ([]*Share, error) {
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

// List fetches all shared secrets for a given username
func (t withTrace) List(ctx context.Context, username string) ([]*Share, error) {
	ctx, s := spanner.Start(ctx, "secret.List")
	defer s.End()
	s.Add(
		attr.String("for_user", username),
	)

	share, err := t.r.List(ctx, username)
	if err != nil {
		s.Event("error listing shared secrets", attr.New("error", err.Error()))
		return nil, err
	}
	return share, nil
}

// ListTarget is similar to List, but returns secrets that are shared with a target user
func (t withTrace) ListTarget(ctx context.Context, target string) ([]*Share, error) {
	ctx, s := spanner.Start(ctx, "secret.ListTarget")
	defer s.End()
	s.Add(
		attr.String("for_user", target),
	)

	share, err := t.r.ListTarget(ctx, target)
	if err != nil {
		s.Event("error listing shared secrets", attr.New("error", err.Error()))
		return nil, err
	}
	return share, nil
}

// Delete removes the user `target` from the secret share
func (t withTrace) Delete(ctx context.Context, sh *Share) error {
	ctx, s := spanner.Start(ctx, "secret.Delete")
	defer s.End()
	s.Add(
		attr.String("from_user", sh.Owner.Username),
	)

	err := t.r.Delete(ctx, sh)
	if err != nil {
		s.Event("error deleting shared secret", attr.New("error", err.Error()))
		return err
	}
	return nil
}
