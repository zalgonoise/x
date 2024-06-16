package iter

import (
	"context"
	"database/sql"
)

// Scannable represents a type that implements the Scan method consuming a pointer to sql.Rows. This item should declare
// Scan with a pointer receiver in order for it to be mutated, and in order for it to be instantiated using the data
// from a sql.Rows scan.
//
// This interface allows direct access to type T as a pointer.
type Scannable[T any] interface {
	Scan(*sql.Rows) error
	*T
}

// QuerierContext describes a type that implements the QueryContext method, such as sql.DB and sql.Tx.
type QuerierContext interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// QueryContext executes a QueryContext call using the input QuerierContext, with input context.Context, query string
// and values arguments. Type T must be a Scannable[T] type implementing a Scan method that consumes *sql.Rows.
//
// This function returns a sequence iterator that yields a pointer to a type T item and an error if raised. The
// resulting iterator returns a boolean on weather the iteration was completed successfully or not.
//
// This is a fail-fast implementation and the call will exit on the first raised error (either on item.Scan, or from
// sql.Rows.Err.
func QueryContext[S Scannable[T], T any](
	ctx context.Context, q QuerierContext, query string, values ...any,
) (Seq[*T, error], error) {
	rows, err := q.QueryContext(ctx, query, values...)
	if err != nil {
		return nil, err
	}

	return func(yield func(*T, error) bool) bool {
		for rows.Next() {
			// initialize a new pointer instance of T, as a Scannable[T] interface type:
			var item S = new(T)

			// as a pointer, item has access to Scan, which mutates it
			if err := item.Scan(rows); err != nil {
				yield(nil, err)

				return false
			}

			// since Scannable[T] is a reference to type T, we're still able to dereference it from *T to T when appending
			if !yield(item, nil) {
				return false
			}
		}

		if err := rows.Close(); err != nil {
			yield(nil, err)

			return false
		}

		if err := rows.Err(); err != nil {
			yield(nil, err)

			return false
		}

		return true
	}, nil
}
