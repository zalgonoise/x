package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/exp/constraints"
)

type Scanner interface {
	Scan(dest ...interface{}) error
}

type RowQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type RowsQuerier interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

type Executer interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type Querier interface {
	RowQuerier
	RowsQuerier
	Executer
}

func ToSQLString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func ToSQLInt64[T constraints.Integer](v T) sql.NullInt64 {
	return sql.NullInt64{Int64: int64(v), Valid: int64(v) >= 0}
}

func ToSQLTime(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: t != time.Time{} && t.Unix() != 0}
}

func IsEntityFound(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDBError, err)
	}
	if n == 0 {
		return ErrNotFoundUser
	}
	return nil
}
