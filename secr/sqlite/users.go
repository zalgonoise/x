package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/zalgonoise/x/secr/user"
)

var (
	ErrDBError      = errors.New("database error")
	ErrNotFoundUser = errors.New("user not found")
)

type dbUser struct {
	ID        sql.NullInt64
	Username  sql.NullString
	Name      sql.NullString
	Hash      sql.NullString
	Salt      sql.NullString
	CreatedAt sql.NullTime
	UpdatedAt sql.NullTime
}

var _ user.Repository = &userRepository{nil}

type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a user.Repository from the SQL DB `db`
func NewUserRepository(db *sql.DB) user.Repository {
	return &userRepository{db}
}

// Create will create a user `u`, returning its ID and an error
func (ur *userRepository) Create(ctx context.Context, u *user.User) (uint64, error) {
	dbu := newDBUser(u)
	res, err := ur.db.ExecContext(ctx, `
INSERT INTO users (username, name, hash, salt)
VALUES (?, ?, ?, ?)
`, dbu.Username, dbu.Name, dbu.Hash, dbu.Salt)

	if err != nil {
		return 0, fmt.Errorf("%w: failed to create user %s: %v", ErrDBError, u.Username, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%w: failed to create user %s: %v", ErrDBError, u.Username, err)
	}
	if id == 0 {
		return 0, fmt.Errorf("%w: user was not created %s", ErrDBError, u.Username)
	}
	return uint64(id), nil
}

// Get returns the user identified by `username`, and an error
func (ur *userRepository) Get(ctx context.Context, username string) (*user.User, error) {
	row := ur.db.QueryRowContext(ctx, `
SELECT u.id, u.username, u.name, u.hash, u.salt, u.created_at, u.updated_at
FROM users AS u
WHERE u.username = ?
	`, ToSQLString(username))

	user, err := ur.scanUser(row)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// List returns all the users, and an error
func (ur *userRepository) List(ctx context.Context) ([]*user.User, error) {
	rows, err := ur.db.QueryContext(ctx, `
SELECT u.id, u.username, u.name, u.created_at, u.updated_at
FROM users AS u
	`)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to list users: %v", ErrDBError, err)
	}

	users, err := ur.scanUsers(rows)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to list users: %v", ErrDBError, err)
	}
	return users, nil
}

// Update will update the user `username` with its updated version `updated`. Returns an error
func (ur *userRepository) Update(ctx context.Context, username string, updated *user.User) error {
	dbu := newDBUser(updated)
	res, err := ur.db.ExecContext(ctx, `
UPDATE users
SET name = ?, hash = ?
WHERE u.username = ?
`, dbu.Name, dbu.Hash, ToSQLString(username))

	if err != nil {
		return fmt.Errorf("%w: failed to update user %s: %v", ErrDBError, username, err)
	}

	err = IsUserFound(res)
	if err != nil {
		return fmt.Errorf("%w: failed to update user %s: %v", ErrDBError, username, err)
	}
	return nil
}

// Delete removes the user identified by `username`, returning an error
func (ur *userRepository) Delete(ctx context.Context, username string) error {
	res, err := ur.db.ExecContext(ctx, `
	DELETE FROM users WHERE username = ?
	`, ToSQLString(username))

	if err != nil {
		return fmt.Errorf("%w: failed to delete user %s: %v", ErrDBError, username, err)
	}

	err = IsUserFound(res)
	if err != nil {
		return fmt.Errorf("%w: failed to delete user %s: %v", ErrDBError, username, err)
	}

	return nil
}

func (ur *userRepository) scanUser(r Scanner) (u *user.User, err error) {
	if r == nil {
		return nil, fmt.Errorf("%w: failed to find this user", ErrNotFoundUser)
	}
	dbu := new(dbUser)
	err = r.Scan(
		&dbu.ID,
		&dbu.Username,
		&dbu.Name,
		&dbu.Hash,
		&dbu.Salt,
		&dbu.CreatedAt,
		&dbu.UpdatedAt,
	)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return nil, ErrNotFoundUser
		}
		return nil, fmt.Errorf("%w: failed to scan DB row: %v", ErrDBError, err)
	}
	return dbu.toDomainEntity(), nil
}

func (ur *userRepository) scanUsers(rs *sql.Rows) ([]*user.User, error) {
	var users = []*user.User{}

	defer rs.Close()
	for rs.Next() {
		u, err := ur.scanUser(rs)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		users = append(users, u)
	}
	return users, nil
}

func (u *dbUser) toDomainEntity() *user.User {
	return &user.User{
		ID:        uint64(u.ID.Int64),
		Username:  u.Username.String,
		Name:      u.Name.String,
		Hash:      u.Hash.String,
		Salt:      u.Salt.String,
		CreatedAt: u.CreatedAt.Time,
		UpdatedAt: u.UpdatedAt.Time,
	}
}

func newDBUser(u *user.User) *dbUser {
	return &dbUser{
		Username: ToSQLString(u.Username),
		Name:     ToSQLString(u.Name),
		Hash:     ToSQLString(u.Hash),
		Salt:     ToSQLString(u.Salt),
	}
}
