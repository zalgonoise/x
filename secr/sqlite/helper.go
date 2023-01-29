package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/zalgonoise/x/errors"
	"golang.org/x/exp/constraints"
)

// Scanner describes an object that scans values into destination pointers
type Scanner interface {
	Scan(dest ...interface{}) error
}

// RowQuerier describes an object that queries a single row in a SQL database
type RowQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// RowsQuerier describes an object that queries multiple rows in a SQL database
type RowsQuerier interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// Executer describes an object that executes a mutation in a SQL database
type Executer interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// Querier encapsulates a RowQuerier, RowsQuerier and Executer implementations
type Querier interface {
	RowQuerier
	RowsQuerier
	Executer
}

// ToSQLString converts the input string into a sql.NullString
func ToSQLString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

// ToSQLInt64 converts the generic integer into a sql.NullInt64
func ToSQLInt64[T constraints.Integer](v T) sql.NullInt64 {
	return sql.NullInt64{Int64: int64(v), Valid: v >= 0}
}

// ToSQLTime converts the input time into a sql.NullTime
func ToSQLTime(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: t != time.Time{} && t.Unix() != 0}
}

// IsUserFound returns an error if the entity is not found
func IsUserFound(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return errors.Join(ErrDBError, err)
	}
	if n == 0 {
		return ErrNotFoundUser
	}
	return nil
}

// IsSecretFound returns an error if the entity is not found
func IsSecretFound(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return errors.Join(ErrDBError, err)
	}
	if n == 0 {
		return ErrNotFoundSecret
	}
	return nil
}

// IsShareFound returns an error if the entity is not found
func IsShareFound(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return errors.Join(ErrDBError, err)
	}
	if n == 0 {
		return ErrNotFoundShare
	}
	return nil
}
