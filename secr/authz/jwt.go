package authz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/zalgonoise/x/secr/user"
)

var (
	ErrMissingExpiry = errors.New("missing expiry claim")
	ErrMissingUser   = errors.New("missing user claim")
	ErrInvalidUser   = errors.New("invalid user claim")
	ErrExpired       = errors.New("token has expired")
	ErrInvalidToken  = errors.New("invalid token")
)

type Authorizer interface {
	NewToken(ctx context.Context, u *user.User) (string, error)
	Refresh(ctx context.Context, u *user.User, token string) (string, error)
	Validate(ctx context.Context, u *user.User, token string) (bool, error)
}

func NewAuthorizer(signingKey []byte) Authorizer {
	return &authz{signingKey}
}

type authz struct {
	signingKey []byte
}

type jwtUser struct {
	Username  string    `json:username`
	Name      string    `json:name`
	CreatedAt time.Time `json:created_at`
	UpdatedAt time.Time `json:updated_at`
}

func (a *authz) NewToken(ctx context.Context, u *user.User) (string, error) {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour * 1).Unix(),
		"user": jwtUser{
			Username:  u.Username,
			Name:      u.Name,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		},
	})

	token, err := tok.SignedString(a.signingKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}
	return token, nil
}

func (a *authz) Refresh(ctx context.Context, u *user.User, token string) (string, error) {
	ok, err := a.Validate(ctx, u, token)
	if err != nil {
		return "", fmt.Errorf("failed to validate token: %v", err)
	}
	if !ok {
		return "", ErrInvalidToken
	}
	return a.NewToken(ctx, u)
}

func (a *authz) Validate(ctx context.Context, u *user.User, token string) (bool, error) {
	tok, err := jwt.Parse(token, a.parseToken)

	if err != nil {
		return false, fmt.Errorf("failed to parse token: %v", err)
	}
	claims := tok.Claims.(jwt.MapClaims)

	exp, ok := claims["exp"]
	if !ok {
		return false, ErrMissingExpiry
	}
	expTime := time.Unix(int64(exp.(float64)), 0)
	if time.Now().After(expTime) {
		return false, ErrExpired
	}
	v, ok := claims["user"]
	if !ok {
		return false, ErrMissingUser
	}
	val := v.(jwtUser)

	if val.Username != u.Username || val.Name != u.Name {
		return false, ErrInvalidUser
	}
	return true, nil
}

func (a *authz) parseToken(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return a.signingKey, nil
}
