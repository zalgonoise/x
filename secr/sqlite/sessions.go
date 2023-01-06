package sqlite

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zalgonoise/x/secr/authz"
	"github.com/zalgonoise/x/secr/keys"
	"github.com/zalgonoise/x/secr/session"
	"github.com/zalgonoise/x/secr/user"
)

var _ session.Repository = &sessionRepository{nil, nil, nil}

type dbPassword struct {
	User sql.NullInt64
	Hash sql.NullString
	Salt sql.NullString
}

type sessionRepository struct {
	db   *sql.DB
	auth authz.Authorizer
	kv   keys.Repository
}

func NewSessionRepository(db *sql.DB, auth authz.Authorizer, k keys.Repository) session.Repository {
	return &sessionRepository{db, auth, k}
}

func (sr *sessionRepository) login(ctx context.Context, u *user.User, password string) (*dbPassword, error) {
	row := sr.db.QueryRowContext(ctx, `
SELECT u.id, u.hash, u.salt
FROM users AS u
WHERE u.username = ?
	`, ToSQLString(u.Username))

	pw, err := scanPassword(row)
	if err != nil {
		return nil, err
	}

	hashedPassword := sha256.Sum256(append([]byte(password), []byte(pw.Salt.String)...))
	if string(hashedPassword[:]) == pw.Hash.String {
		return pw, nil
	}

	// try to use JWT as password
	ok, err := sr.auth.Validate(ctx, u, password)
	if err != nil {
		if errors.Is(authz.ErrExpired, err) {
			derr := sr.kv.Delete(ctx, u.Username, keys.TokenKey)
			if derr != nil {
				err = fmt.Errorf("%w: failed to remove old token: %v", err, derr)
			}
		}
		return nil, fmt.Errorf("failed to validate JWT: %v", err)
	}
	if !ok {
		return nil, fmt.Errorf("%w: couldn't login with username %s", ErrIncorrectPassword, u.Username)
	}
	return pw, nil
}

func (sr *sessionRepository) Login(ctx context.Context, u *user.User, password string) (*session.Session, error) {
	_, err := sr.login(ctx, u, password)
	if err != nil {
		return nil, err
	}

	token, err := sr.auth.NewToken(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new session token: %v", err)
	}

	err = sr.kv.Set(ctx, u.Username, keys.TokenKey, []byte(token))
	if err != nil {
		return nil, fmt.Errorf("failed to store the new session token: %v", err)
	}

	return &session.Session{
		User:  *u,
		Token: token,
	}, nil
}
func (sr *sessionRepository) Logout(ctx context.Context, u *user.User) error {
	err := sr.kv.Delete(ctx, u.Username, keys.TokenKey)
	if err != nil {
		return fmt.Errorf("failed to log user out: %v", err)
	}
	return nil
}
func (sr *sessionRepository) ChangePassword(ctx context.Context, u *user.User, password, newPassword string) error {
	pw, err := sr.login(ctx, u, password)
	if err != nil {
		return err
	}

	hashedPassword := sha256.Sum256(append([]byte(newPassword), []byte(pw.Salt.String)...))

	res, err := sr.db.ExecContext(ctx, `
UPDATE users
SET hash = ?
WHERE u.username = ?
`, ToSQLString(string(hashedPassword[:])), ToSQLString(u.Username))

	if err != nil {
		return fmt.Errorf("%w: failed to update user's password: %v", ErrDBError, err)
	}
	if err := IsEntityFound(res); err != nil {
		return fmt.Errorf("%w: failed to update user's password: %v", ErrDBError, err)
	}

	return nil
}
func (sr *sessionRepository) Refresh(ctx context.Context, u *user.User, token string) (*session.Session, error) {
	newToken, err := sr.auth.Refresh(ctx, u, token)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %v", err)
	}

	err = sr.kv.Set(ctx, u.Username, keys.TokenKey, []byte(newToken))
	if err != nil {
		return nil, fmt.Errorf("failed to store the new session token: %v", err)
	}

	return &session.Session{
		User:  *u,
		Token: newToken,
	}, nil
}

func scanPassword(r *sql.Row) (*dbPassword, error) {
	if r == nil {
		return nil, fmt.Errorf("%w: failed to find this user", ErrNotFoundUser)
	}

	var pw = new(dbPassword)
	err := r.Scan(
		&pw.User,
		&pw.Hash,
		&pw.Salt,
	)

	if err != nil {
		return nil, fmt.Errorf("%w: failed to scan row: %v", ErrDBError, err)
	}

	return pw, nil
}
