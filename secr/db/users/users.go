package users

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"

	"github.com/zalgonoise/x/secr/user"
)

var (
	ErrDBError           = errors.New("database error")
	ErrIncorrectPassword = errors.New("incorrect username or password")
	ErrNotFoundUser      = errors.New("user not found")
	ErrAlreadyExistsUser = errors.New("user already exists")
	ErrNoUser            = errors.New("no user provided")
	ErrNoPassword        = errors.New("no password provided")
	ErrNoName            = errors.New("no name provided")
)

var _ user.Repository = &userRepository{nil}

type dbUser struct {
	ID        sql.NullInt64
	Username  sql.NullString
	Name      sql.NullString
	Hash      sql.NullString
	Salt      sql.NullString
	CreatedAt sql.NullTime
	UpdatedAt sql.NullTime
}

type userRepository struct {
	db *sql.DB
}

func (ur *userRepository) Create(ctx context.Context, username, password, name string) (uint64, error) {
	if username == "" {
		return 0, fmt.Errorf("%w: username cannot be empty", ErrNoUser)
	}
	if password == "" {
		return 0, fmt.Errorf("%w: password cannot be empty", ErrNoPassword)
	}
	if name == "" {
		return 0, fmt.Errorf("%w: name cannot be empty", ErrNoName)
	}

	salt := saltGen.NewSalt()
	hashedPassword := sha256.Sum256(append([]byte(password), salt[:]...))

	res, err := ur.db.ExecContext(ctx, `
INSERT INTO users (username, name, hash, salt)
VALUES (?, ?, ?, ?)
`, username, name, hashedPassword, string(salt[:]))

	if err != nil {
		return 0, fmt.Errorf("%w: failed to create user %s: %v", ErrDBError, username, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%w: failed to create user %s: %v", ErrDBError, username, err)
	}
	if id == 0 {
		return 0, fmt.Errorf("%w: user was not created %s", ErrDBError, username)
	}
	return uint64(id), nil
}
func (ur *userRepository) Update(ctx context.Context, username string, updated *user.User) error {
	if username == "" {
		return fmt.Errorf("%w: username cannot be empty", ErrNoUser)
	}
	if updated.Name == "" {
		return fmt.Errorf("%w: new name cannot be empty", ErrNoName)
	}
	u := newDBUser(updated)

	res, err := ur.db.ExecContext(ctx, `
UPDATE users
SET name = ?
WHERE u.username = ?
`, u.Name, ToSQLString(username))

	if err != nil {
		return fmt.Errorf("%w: failed to update user %s: %v", ErrDBError, username, err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: failed to update user %s: %v", ErrDBError, username, err)
	}
	if n == 0 {
		return fmt.Errorf("%w: user was not updated %s", ErrDBError, username)
	}

	return nil
}
func (ur *userRepository) Get(ctx context.Context, username string) (*user.User, error) {
	if username == "" {
		return nil, fmt.Errorf("%w: username cannot be empty", ErrNoUser)
	}

	row := ur.db.QueryRowContext(ctx, `
SELECT u.id, u.username, u.name, u.hash, u.salt, u.created_at, u.updated_at
FROM users AS u
WHERE u.username = ?
	`, username)

	user, err := ur.scanUser(row)
	if err != nil {
		return nil, err
	}

	return user, nil
}
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

func (ur *userRepository) Delete(ctx context.Context, username string) error {
	if username == "" {
		return fmt.Errorf("%w: username cannot be empty", ErrNoUser)
	}

	res, err := ur.db.ExecContext(ctx, `
	DELETE u
	FROM users AS u
	WHERE u.username = ?
	`, username)

	if err != nil {
		return fmt.Errorf("%w: failed to delete user %s: %v", ErrDBError, username, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: failed to delete user %s: %v", ErrDBError, username, err)
	}
	if n == 0 {
		return fmt.Errorf("%w: user was not deleted %s", ErrDBError, username)
	}

	return nil
}

func (ur *userRepository) scanUser(r Scanner) (u *user.User, err error) {
	if r == nil {
		return nil, fmt.Errorf("%w: failed to find this user", ErrNotFoundUser)
	}
	dbu := new(dbUser)
	err = r.Scan(
		&u.ID,
		&u.Username,
		&u.Name,
		&u.Hash,
		&u.Salt,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to scan DB row: %v", ErrDBError, err)
	}
	return toDomainEntity(dbu), nil
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

func toDomainEntity(u *dbUser) *user.User {
	return &user.User{
		ID:        uint64(u.ID.Int64),
		Username:  u.Username.String,
		Name:      u.Name.String,
		Hash:      u.Hash.String,
		Salt:      u.Hash.String,
		CreatedAt: u.CreatedAt.Time,
		UpdatedAt: u.UpdatedAt.Time,
	}
}

func newDBUser(u *user.User) *dbUser {
	return &dbUser{
		ID:       ToSQLInt64(u.ID),
		Username: ToSQLString(u.Username),
		Name:     ToSQLString(u.Name),
		Hash:     ToSQLString(u.Hash),
		Salt:     ToSQLString(u.Salt),
	}
}
