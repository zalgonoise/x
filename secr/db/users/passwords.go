package users

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zalgonoise/x/secr/session"
)

var _ session.Repository = &sessionRepository{nil}

type dbPassword struct {
	User sql.NullInt64
	Hash sql.NullString
	Salt sql.NullString
}
type dbSession struct {
	User  uint64
	Token string
}

type sessionRepository struct {
	db *sql.DB
}

func (sr *sessionRepository) login(ctx context.Context, username, password string) (*dbPassword, error) {
	row := sr.db.QueryRowContext(ctx, `
SELECT u.id, u.hash, u.salt
FROM users AS u
WHERE u.username = ?
	`, username)

	pw, err := scanPassword(row)
	if err != nil {
		return nil, err
	}

	hashedPassword := sha256.Sum256(append([]byte(password), []byte(pw.Salt.String)...))
	if string(hashedPassword[:]) != pw.Hash.String {
		return nil, fmt.Errorf("%w: couldn't login with username %s", ErrIncorrectPassword, username)
	}
	return pw, nil
}

func (sr *sessionRepository) Login(ctx context.Context, username, password string) (*session.Session, error) {
	_, err := sr.login(ctx, username, password)
	if err != nil {
		return nil, err
	}

	// TODO: create token and return it

	return nil, nil
}
func (sr *sessionRepository) Logout(ctx context.Context, username string) error {
	// TODO: delete the user's active token
	return nil
}
func (sr *sessionRepository) ChangePassword(ctx context.Context, username, password, newPassword string) error {
	pw, err := sr.login(ctx, username, password)
	if err != nil {
		return err
	}

	hashedPassword := sha256.Sum256(append([]byte(newPassword), []byte(pw.Salt.String)...))

	res, err := sr.db.ExecContext(ctx, `
UPDATE users
SET hash = ?
WHERE u.username = ?
`, ToSQLString(string(hashedPassword[:])), ToSQLString(username))

	if err != nil {
		return fmt.Errorf("%w: failed to update user's password: %v", ErrDBError, err)
	}
	if err := IsEntityFound(res); err != nil {
		return fmt.Errorf("%w: failed to update user's password: %v", ErrDBError, err)
	}

	return nil
}
func (sr *sessionRepository) Refresh(ctx context.Context, username, token string) (*session.Session, error) {
	return nil, nil
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
