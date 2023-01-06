package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/zalgonoise/x/secr/keys"
	"github.com/zalgonoise/x/secr/secret"
)

var _ secret.Repository = &secretRepository{nil, nil}

var (
	ErrNoKey          = errors.New("no secret key provided")
	ErrNoValue        = errors.New("no secret value provided")
	ErrNotFoundSecret = errors.New("secret not found")
)

type dbSecret struct {
	ID        sql.NullInt64
	Name      sql.NullString
	Value     sql.NullString
	CreatedAt sql.NullTime
}

type secretRepository struct {
	db *sql.DB
	kv keys.Repository
}

func NewSecretRepository(db *sql.DB, kv keys.Repository) secret.Repository {
	return &secretRepository{db, kv}
}

func newDBSecret(s *secret.Secret) *dbSecret {
	return &dbSecret{
		Name: ToSQLString(s.Key),
	}
}

func (sr *secretRepository) Create(ctx context.Context, username string, s *secret.Secret) error {
	if username == "" {
		return fmt.Errorf("%w: username cannot be empty", ErrNoUser)
	}
	if s.Key == "" {
		return fmt.Errorf("%w: key cannot be empty", ErrNoKey)
	}
	if len(s.Value) == 0 {
		return fmt.Errorf("%w: value cannot be empty", ErrNoValue)
	}

	// create secret in bbolt
	// will go to service when impl
	err := sr.kv.Set(ctx, username, s.Key, s.Value)
	if err != nil {
		return err
	}
	rollback := func() {
		sr.kv.Delete(ctx, username, s.Key)
	}

	dbs := newDBSecret(s)

	res, err := sr.db.ExecContext(ctx, `
INSERT INTO secrets (user_id, name)
VALUES (
	(SELECT u.id FROM users AS u WHERE u.username = ?), 
	?, ?)
`, username, dbs.Name, dbs.Value)

	if err != nil {
		rollback()
		return fmt.Errorf("%w: failed to create secret %s: %v", ErrDBError, s.Key, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		rollback()
		return fmt.Errorf("%w: failed to create secret %s: %v", ErrDBError, s.Key, err)
	}
	if id == 0 {
		rollback()
		return fmt.Errorf("%w: secret was not created %s", ErrDBError, s.Key)
	}
	return nil
}

func (sr *secretRepository) Get(ctx context.Context, username string, key string) (*secret.Secret, error) {
	if username == "" {
		return nil, fmt.Errorf("%w: username cannot be empty", ErrNoUser)
	}
	if key == "" {
		return nil, fmt.Errorf("%w: key cannot be empty", ErrNoKey)
	}

	// will go to service when impl
	scr, err := sr.kv.Get(ctx, username, key)
	if err != nil {
		return nil, err
	}

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

	s.Value = scr

	return s, nil

}
func (sr *secretRepository) List(ctx context.Context, username string) ([]*secret.Secret, error) {
	if username == "" {
		return nil, fmt.Errorf("%w: username cannot be empty", ErrNoUser)
	}

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
	for _, s := range secrets {
		s.Value, err = sr.kv.Get(ctx, username, s.Key)
		if err != nil {
			return nil, err
		}
	}

	return secrets, nil
}
func (sr *secretRepository) Delete(ctx context.Context, username string, key string) error {
	if username == "" {
		return fmt.Errorf("%w: username cannot be empty", ErrNoUser)
	}
	if key == "" {
		return fmt.Errorf("%w: secret key cannot be empty", ErrNoKey)
	}

	// will go to service when impl
	s, err := sr.kv.Get(ctx, username, key)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrNotFoundSecret, err)
	}
	err = sr.kv.Delete(ctx, username, key)
	if err != nil {
		return err
	}
	rollback := func() {
		sr.kv.Set(ctx, username, key, s)
	}

	res, err := sr.db.ExecContext(ctx, `
	DELETE s
	FROM secrets AS s
		JOIN users AS u ON u.id = s.user_id
	WHERE u.username = ? 
		AND s.name = ?
	`, username)

	if err != nil {
		rollback()
		return fmt.Errorf("%w: failed to delete secret %s: %v", ErrDBError, key, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		rollback()
		return fmt.Errorf("%w: failed to delete secret %s: %v", ErrDBError, key, err)
	}
	if n == 0 {
		rollback()
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
		Value:     []byte(s.Value.String),
		CreatedAt: s.CreatedAt.Time,
	}
}
