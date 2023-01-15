package user

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

// Create will create a user `u`, returning its ID and an error
func (t withTrace) Create(ctx context.Context, u *User) (uint64, error) {
	ctx, s := spanner.Start(ctx, "user.Create")
	defer s.End()
	s.Add(
		attr.String("username", u.Username),
	)

	id, err := t.r.Create(ctx, u)
	if err != nil {
		s.Event("error creating user", attr.New("error", err.Error()))
		return id, err
	}
	return id, err
}

// Update will update the user `username` with its updated version `updated`. Returns an error
func (t withTrace) Update(ctx context.Context, username string, updated *User) error {
	ctx, s := spanner.Start(ctx, "user.Update")
	defer s.End()
	s.Add(
		attr.String("username", username),
		attr.String("new_username", updated.Username),
	)

	err := t.r.Update(ctx, username, updated)
	if err != nil {
		s.Event("error updating user", attr.New("error", err.Error()))
		return err
	}
	return err
}

// Get returns the user identified by `username`, and an error
func (t withTrace) Get(ctx context.Context, username string) (*User, error) {
	ctx, s := spanner.Start(ctx, "user.Get")
	defer s.End()
	s.Add(
		attr.String("username", username),
	)

	u, err := t.r.Get(ctx, username)
	if err != nil {
		s.Event("error fetching user", attr.New("error", err.Error()))
		return u, err
	}
	return u, err
}

// List returns all the users, and an error
func (t withTrace) List(ctx context.Context) ([]*User, error) {
	ctx, s := spanner.Start(ctx, "user.List")
	defer s.End()

	u, err := t.r.List(ctx)
	if err != nil {
		s.Event("error listing users", attr.New("error", err.Error()))
		return u, err
	}
	return u, err
}

// Delete removes the user identified by `username`, returning an error
func (t withTrace) Delete(ctx context.Context, username string) error {
	ctx, s := spanner.Start(ctx, "user.Delete")
	defer s.End()
	s.Add(
		attr.String("username", username),
	)

	err := t.r.Delete(ctx, username)
	if err != nil {
		s.Event("error deleting user", attr.New("error", err.Error()))
		return err
	}
	return err
}
