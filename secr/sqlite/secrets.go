package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/zalgonoise/x/secr/secret"
)

var _ secret.Repository = &secretRepository{nil}

var (
	ErrNotFoundSecret = errors.New("secret not found")
)

type dbSecret struct {
	ID        sql.NullInt64
	Name      sql.NullString
	CreatedAt sql.NullTime
}

type secretRepository struct {
	db *sql.DB
}

// NewSecretRepository creates a secret.Repository from the SQL DB `db`
func NewSecretRepository(db *sql.DB) secret.Repository {
	return &secretRepository{db}
}

func newDBSecret(s *secret.Secret) *dbSecret {
	return &dbSecret{
		Name: ToSQLString(s.Key),
	}
}

// Create will create (or overwrite) the secret identified by `s.Key`, for user `username`,
// returning an error
func (sr *secretRepository) Create(ctx context.Context, username string, s *secret.Secret) (uint64, error) {
	dbs := newDBSecret(s)
	res, err := sr.db.ExecContext(ctx, `
INSERT INTO secrets (user_id, name)
VALUES (
	(SELECT u.id FROM users AS u WHERE u.username = ?), 
	?)
`, username, dbs.Name)

	if err != nil {
		return 0, fmt.Errorf("%w: failed to create secret %s: %v", ErrDBError, s.Key, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%w: failed to create secret %s: %v", ErrDBError, s.Key, err)
	}
	if id == 0 {
		return 0, fmt.Errorf("%w: secret was not created %s", ErrDBError, s.Key)
	}

	return uint64(id), nil
}

// Get fetches a secret identified by `key` for user `username`. Returns a secret and an error
func (sr *secretRepository) Get(ctx context.Context, username string, key string) (*secret.Secret, error) {
	row := sr.db.QueryRowContext(ctx, `
SELECT s.id, s.name, s.created_at
FROM secrets AS s
	JOIN users AS u ON u.id = s.user_id
WHERE u.username = ?
	AND s.name = ?
	`, ToSQLString(username), ToSQLString(key))

	s, err := sr.scanSecret(row)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// List returns all secrets belonging to user `username`, and an error
func (sr *secretRepository) List(ctx context.Context, username string) ([]*secret.Secret, error) {
	rows, err := sr.db.QueryContext(ctx, `
SELECT s.id, s.name, s.created_at
FROM secrets AS s
	JOIN users AS u ON u.id = s.user_id
WHERE u.username = ?
		`, ToSQLString(username))

	if err != nil {
		return nil, fmt.Errorf("%w: failed to list secrets: %v", ErrDBError, err)
	}

	secrets, err := sr.scanSecrets(rows)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to list secrets: %v", ErrDBError, err)
	}

	return secrets, nil
}

// Delete removes the secret identified by `key`, for user `username`. Returns an error
func (sr *secretRepository) Delete(ctx context.Context, username string, key string) error {
	res, err := sr.db.ExecContext(ctx, `
	DELETE s
	FROM secrets AS s
		JOIN users AS u ON u.id = s.user_id
	WHERE u.username = ? 
		AND s.name = ?
	`, username)

	if err != nil {
		return fmt.Errorf("%w: failed to delete secret %s: %v", ErrDBError, key, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: failed to delete secret %s: %v", ErrDBError, key, err)
	}
	if n == 0 {
		return fmt.Errorf("%w: secret was not deleted %s", ErrDBError, key)
	}

	return nil
}

func (sr *secretRepository) scanSecret(r Scanner) (s *secret.Secret, err error) {
	if r == nil {
		return nil, fmt.Errorf("%w: failed to find this secret", ErrNotFoundSecret)
	}
	dbs := new(dbSecret)
	err = r.Scan(
		&dbs.ID,
		&dbs.Name,
		&dbs.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to scan DB row: %v", ErrDBError, err)
	}
	return dbs.toDomainEntity(), nil
}

func (sr *secretRepository) scanSecrets(rs *sql.Rows) ([]*secret.Secret, error) {
	var secrets = []*secret.Secret{}

	defer rs.Close()
	for rs.Next() {
		s, err := sr.scanSecret(rs)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		secrets = append(secrets, s)
	}
	return secrets, nil
}

func (s *dbSecret) toDomainEntity() *secret.Secret {
	return &secret.Secret{
		Key:       s.Name.String,
		CreatedAt: s.CreatedAt.Time,
	}
}
