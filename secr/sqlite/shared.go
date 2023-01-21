package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/zalgonoise/x/secr/secret"
	"github.com/zalgonoise/x/secr/shared"
	"github.com/zalgonoise/x/secr/user"
)

var (
	ErrNotFoundShare = errors.New("shared secret not found")
)

var _ shared.Repository = &sharedRepository{nil}

type dbShare struct {
	ID        sql.NullInt64
	Secret    sql.NullString
	Owner     sql.NullString
	Target    sql.NullString
	Until     sql.NullTime
	CreatedAt sql.NullTime
}

type sharedRepository struct {
	db *sql.DB
}

// NewSharedRepository creates a shared.Repository from the SQL DB `db`
func NewSharedRepository(db *sql.DB) shared.Repository {
	return &sharedRepository{db}
}

// Get fetches the secret's share metadata for a given username and secret key
func (sr *sharedRepository) Get(ctx context.Context, username, secretName string) ([]*shared.Share, error) {
	rows, err := sr.db.QueryContext(ctx, `
SELECT s.id, x.name, o.username, t.username, s.until, s.created_at
FROM shared_secrets AS s
	JOIN users AS o ON o.id = s.owner_id
	JOIN users AS t ON t.id = s.shared_with
	JOIN secrets AS x ON x.id = s.secret_id
WHERE o.username = ?
	AND x.name = ?
`, ToSQLString(username), ToSQLString(secretName))

	if err != nil {
		return nil, fmt.Errorf("%w: failed to list shared secrets: %v", ErrDBError, err)
	}

	shares, err := sr.scanShares(rows)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to list shared secrets: %v", ErrDBError, err)
	}
	return shares, nil
}

func (sr *sharedRepository) List(ctx context.Context, username string) ([]*shared.Share, error) {
	rows, err := sr.db.QueryContext(ctx, `
SELECT s.id, x.name, o.username, t.username, s.until, s.created_at
FROM shared_secrets AS s
	JOIN users AS o ON o.id = s.owner_id
	JOIN users AS t ON t.id = s.shared_with
	JOIN secrets AS x ON x.id = s.secret_id
WHERE o.username = ?
`, ToSQLString(username))

	if err != nil {
		return nil, fmt.Errorf("%w: failed to list shared secrets: %v", ErrDBError, err)
	}

	shares, err := sr.scanShares(rows)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to list shared secrets: %v", ErrDBError, err)
	}
	return shares, nil
}

// ListTarget is similar to List, but returns secrets that are shared with a target user
func (sr *sharedRepository) ListTarget(ctx context.Context, target string) ([]*shared.Share, error) {
	rows, err := sr.db.QueryContext(ctx, `
	SELECT s.id, x.name, o.username, t.username, s.until, s.created_at
	FROM shared_secrets AS s
		JOIN users AS o ON o.id = s.owner_id
		JOIN users AS t ON t.id = s.shared_with
		JOIN secrets AS x ON x.id = s.secret_id
	WHERE t.username = ?
	`, ToSQLString(target))

	if err != nil {
		return nil, fmt.Errorf("%w: failed to list shared secrets: %v", ErrDBError, err)
	}

	shares, err := sr.scanShares(rows)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to list shared secrets: %v", ErrDBError, err)
	}
	return shares, nil
}

// Create shares the secret identified by `secretName`, owned by `owner`, with
// user `target`. Returns an error
func (sr *sharedRepository) Create(ctx context.Context, sh *shared.Share) (uint64, error) {
	shares := newDBShare(sh)
	tx, err := sr.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %v", err)
	}

	var lastID uint64

	for _, dbs := range shares {
		res, err := tx.ExecContext(ctx, `
		INSERT INTO shared_secrets (owner_id, secret_id, shared_with)
		VALUES (
			(SELECT id FROM users WHERE username = ?),
			(SELECT id FROM secrets WHERE name = ?),
			(SELECT id FROM users WHERE username = ?)
		)
		`, dbs.Owner, dbs.Secret, dbs.Target)

		if err != nil {
			_ = tx.Rollback()
			return 0, err
		}
		id, err := res.LastInsertId()
		if err != nil {
			_ = tx.Rollback()
			return 0, err
		}
		if id == 0 {
			_ = tx.Rollback()
			return 0, errors.New("id zero")
		}
		lastID = uint64(id)
	}

	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("failed to create shared secret: %v", err)
	}
	return lastID, nil
}

// Delete removes the user `target` from the secret share
func (sr *sharedRepository) Delete(ctx context.Context, sh *shared.Share) error {
	dbs := newDBShare(sh)
	tx, err := sr.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	for _, share := range dbs {
		res, err := tx.ExecContext(ctx, `
DELETE s
FROM shared_secrets AS s
	JOIN users AS o ON o.id = s.owner_id
	JOIN users AS t ON t.id = s.shared_with
	JOIN secrets AS x ON x.id = s.secret_id
WHERE o.username = ?
	AND x.name = ?
	AND t.username = ?`,
			share.Owner, share.Secret, share.Target)

		if err != nil {
			tx.Rollback()
			return fmt.Errorf("%w: failed to delete shared secret: %v", ErrDBError, err)
		}
		n, err := res.RowsAffected()
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("%w: failed to delete shared secret: %v", ErrDBError, err)
		}
		if n == 0 {
			tx.Rollback()
			return fmt.Errorf("%w: shared secret was not deleted", ErrDBError)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("%w: shared secret was not deleted: %v", ErrDBError, err)
	}
	return nil
}

func toDomainShare(shares ...*dbShare) []*shared.Share {
	if len(shares) == 0 {
		return nil
	}

	s := make([]*shared.Share, 0, len(shares))
	s = append(s, &shared.Share{
		ID: uint64(shares[0].ID.Int64),
		Secret: secret.Secret{
			Key: shares[0].Secret.String,
		},
		Owner: user.User{
			Username: shares[0].Owner.String,
		},
		Target: []user.User{{
			Username: shares[0].Target.String,
		}},
		Until:     shares[0].Until.Time,
		CreatedAt: shares[0].CreatedAt.Time,
	})

	if len(shares) == 1 {
		return s
	}

inputLoop:
	for i := 1; i < len(shares); i++ {
		for _, sh := range s {
			if shares[i].Secret.String == sh.Secret.Key {
				if shares[i].Until.Time == sh.Until {
					for _, t := range sh.Target {
						if t.Username == shares[i].Target.String {
							continue
						}
						sh.Target = append(sh.Target, user.User{
							Username: shares[i].Target.String,
						})
					}
					continue inputLoop
				}
			}
		}
		s = append(s, &shared.Share{
			ID: uint64(shares[i].ID.Int64),
			Secret: secret.Secret{
				Key: shares[i].Secret.String,
			},
			Owner: user.User{
				Username: shares[i].Owner.String,
			},
			Target: []user.User{{
				Username: shares[i].Target.String,
			}},
			Until:     shares[i].Until.Time,
			CreatedAt: shares[i].CreatedAt.Time,
		})
	}

	return s
}

func (sr *sharedRepository) scanShare(r Scanner) (dbs *dbShare, err error) {
	if r == nil {
		return nil, fmt.Errorf("%w: failed to find this share", ErrNotFoundShare)
	}
	dbs = new(dbShare)
	err = r.Scan(
		&dbs.ID,
		&dbs.Secret,
		&dbs.Owner,
		&dbs.Target,
		&dbs.Until,
		&dbs.CreatedAt,
	)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return nil, ErrNotFoundShare
		}
		return nil, fmt.Errorf("%w: failed to scan DB row: %v", ErrDBError, err)
	}
	return dbs, nil
}

func (sr *sharedRepository) scanShares(rs *sql.Rows) ([]*shared.Share, error) {
	var shares []*dbShare

	defer rs.Close()
	for rs.Next() {
		dbs, err := sr.scanShare(rs)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		shares = append(shares, dbs)
	}

	return toDomainShare(shares...), nil
}

func newDBShare(s *shared.Share) []*dbShare {
	shares := make([]*dbShare, 0, len(s.Target))

	for _, t := range s.Target {
		shares = append(shares, &dbShare{
			ID:     ToSQLInt64(s.ID),
			Owner:  ToSQLString(s.Owner.Username),
			Secret: ToSQLString(s.Secret.Key),
			Target: ToSQLString(t.Username),
		})
	}
	return shares
}

func newDBShareQuery(username, secretKey string) *dbShare {
	return &dbShare{
		Owner:  ToSQLString(username),
		Secret: ToSQLString(secretKey),
	}
}
