package iter_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zalgonoise/x/iter"
	_ "modernc.org/sqlite"
)

type testID struct {
	id sql.NullString
}

func (i *testID) Scan(r *sql.Rows) error {
	return r.Scan(&i.id)
}

func TestQueryContext(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		input []string
	}{
		{
			name:  "single_result",
			input: []string{"alpha"},
		},
		{
			name:  "zero_results",
			input: []string{},
		},
		{
			name:  "many_results",
			input: []string{"alpha", "beta", "gamma"},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			ctx := context.Background()
			db := newDB(ctx, t)

			defer db.Close()

			tx, err := db.BeginTx(ctx, nil)
			require.NoError(t, err)

			defer tx.Rollback()

			for i := range testcase.input {
				_, err := tx.ExecContext(ctx, `INSERT INTO testdata (id) VALUES (?)`, testcase.input[i])
				require.NoError(t, err)
			}

			require.NoError(t, tx.Commit())

			seq, err := iter.QueryContext[*testID](ctx, db, `SELECT id FROM testdata`)
			require.NoError(t, err)

			counter := 0
			if !seq(func(id *testID, err error) bool {
				require.NoError(t, err)
				require.Equal(t, id.id.String, testcase.input[counter])
				counter++

				return true
			}) {
				t.Fail()
			}

			require.Equal(t, len(testcase.input), counter)
		})
	}
}

func newDB(ctx context.Context, t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", "file::memory:?_readonly=true&_txlock=immediate&cache=shared")
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
CREATE TABLE testdata (
    id          		TEXT PRIMARY KEY    NOT NULL
) STRICT;`)
	require.NoError(t, err)

	return db
}
