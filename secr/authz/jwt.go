package authz

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/zalgonoise/x/secr/user"
)

type usernameIdentifier string

const ContextUsername usernameIdentifier = "secr:username"

var (
	ErrMissingExpiry = errors.New("missing expiry claim")
	ErrMissingUser   = errors.New("missing user claim")
	ErrInvalidUser   = errors.New("invalid user claim")
	ErrExpired       = errors.New("token has expired")
	ErrInvalidToken  = errors.New("invalid token")
)

// Authorizer is responsible for generating, refreshing and validating JWT
type Authorizer interface {
	// NewToken returns a new JWT for the user `u`, and an error
	NewToken(ctx context.Context, u *user.User) (string, error)
	// Refresh returns a new JWT for the user `u` based on token `token`, and an error
	Refresh(ctx context.Context, u *user.User, token string) (string, error)
	// Validate verifies if the JWT `token` is valid for the user `u`, returning a
	// boolean and an error
	Validate(ctx context.Context, u *user.User, token string) (bool, error)
	// Parse returns the data from a valid JWT
	Parse(ctx context.Context, token string) (*user.User, error)
}

// NewAuthorizer initializes an Authorizer with the signing key `signingKey`
func NewAuthorizer(signingKey []byte) Authorizer {
	return &authz{signingKey}
}

type authz struct {
	signingKey []byte
}

type jwtUser struct {
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewToken returns a new JWT for the user `u`, and an error
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

// Refresh returns a new JWT for the user `u` based on token `token`, and an error
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

// Validate verifies if the JWT `token` is valid for the user `u`, returning a
// boolean and an error
func (a *authz) Validate(ctx context.Context, u *user.User, token string) (bool, error) {
	jwtUser, err := a.Parse(ctx, token)
	if err != nil {
		return false, err
	}

	if jwtUser.Username != u.Username {
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

// Parse returns the data from a valid JWT
func (a *authz) Parse(ctx context.Context, token string) (*user.User, error) {
	tok, err := jwt.Parse(token, a.parseToken)

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}
	claims := tok.Claims.(jwt.MapClaims)

	exp, ok := claims["exp"]
	if !ok {
		return nil, ErrMissingExpiry
	}
	expTime := time.Unix(int64(exp.(float64)), 0)
	if time.Now().After(expTime) {
		return nil, ErrExpired
	}
	v, ok := claims["user"]
	if !ok {
		return nil, ErrMissingUser
	}

	valmap := v.(map[string]interface{})
	vUsername := valmap["username"].(string)
	vName := valmap["name"].(string)

	return &user.User{
		Name:     vName,
		Username: vUsername,
	}, nil
}

// SignRequest sets the input username `u` as a ContextUsername context value for
// the HTTP Request `r`'s context
func SignRequest(u string, r *http.Request) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), ContextUsername, u))
}

// GetCaller returns the username associated with the HTTP Request `r`, as extracted
// from the request's context, under its ContextUsername value (if existing).
//
// Returns the username and an OK-boolean.
func GetCaller(r *http.Request) (string, bool) {
	v := r.Context().Value(ContextUsername)
	if v == nil {
		return "", false
	}
	if u, ok := v.(string); ok {
		return u, true
	}
	return "", false
}
