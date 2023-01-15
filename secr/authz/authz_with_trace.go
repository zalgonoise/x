package authz

import (
	"context"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/spanner"
	"github.com/zalgonoise/x/secr/user"
)

type withTrace struct {
	r Authorizer
}

func WithTrace(r Authorizer) Authorizer {
	return withTrace{
		r: r,
	}
}

// NewToken returns a new JWT for the user `u`, and an error
func (t withTrace) NewToken(ctx context.Context, u *user.User) (string, error) {
	ctx, s := spanner.Start(ctx, "authz.NewToken")
	defer s.End()
	s.Add(
		attr.String("username", u.Username),
	)

	token, err := t.r.NewToken(ctx, u)
	if err != nil {
		s.Event("error creating token for user", attr.New("error", err.Error()))
		return token, err
	}
	return token, nil
}

// Refresh returns a new JWT for the user `u` based on token `token`, and an error
func (t withTrace) Refresh(ctx context.Context, u *user.User, token string) (string, error) {
	ctx, s := spanner.Start(ctx, "authz.Refresh")
	defer s.End()
	s.Add(
		attr.String("username", u.Username),
	)

	token, err := t.r.Refresh(ctx, u, token)
	if err != nil {
		s.Event("error refreshing token for user", attr.New("error", err.Error()))
		return token, err
	}
	return token, nil
}

// Validate verifies if the JWT `token` is valid for the user `u`, returning a
// boolean and an error
func (t withTrace) Validate(ctx context.Context, u *user.User, token string) (bool, error) {
	ctx, s := spanner.Start(ctx, "authz.Validate")
	defer s.End()
	s.Add(
		attr.String("username", u.Username),
	)

	ok, err := t.r.Validate(ctx, u, token)
	if err != nil {
		s.Event("error validating token for user", attr.New("error", err.Error()))
		return ok, err
	}
	return ok, nil
}

// Parse returns the data from a valid JWT
func (t withTrace) Parse(ctx context.Context, token string) (*user.User, error) {
	ctx, s := spanner.Start(ctx, "authz.Parse")
	defer s.End()

	u, err := t.r.Parse(ctx, token)
	if err != nil {
		s.Event("error parsing token", attr.New("error", err.Error()))
		return u, err
	}
	s.Add(
		attr.String("username", u.Username),
	)

	return u, nil
}
