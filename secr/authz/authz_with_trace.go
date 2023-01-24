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
