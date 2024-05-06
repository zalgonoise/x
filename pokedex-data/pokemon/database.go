package pokemon

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/jackc/pgx/v5/stdlib" // Postgres driver
)

func buildInsert(summaries []Summary) (string, []any, error) {
	b := sq.Insert("items").Columns("id", "image_source", "name")

	for i := range summaries {
		b = b.Values(strconv.Itoa(summaries[i].ID), summaries[i].Sprite, summaries[i].Name)
	}

	return b.PlaceholderFormat(sq.Dollar).ToSql()
}

func Load(ctx context.Context, db *sql.DB, summaries []Summary) error {
	q, args, err := buildInsert(summaries)
	if err != nil {
		return err
	}

	res, err := db.ExecContext(ctx, q, args...)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n != int64(len(summaries)) {
		return fmt.Errorf("unexpected number of affected rows: %d", n)
	}

	return nil
}

// typeCast decorates a squirrel.Dollar squirrel.PlaceholderFormat, that also replaces certain
// placeholders with a type-cast one (e.g. $1 --> $1::uuid) using a simple strings.Replacer approach.
type typeCast struct {
	oldnew []string
}

// ReplacePlaceholders implements squirrel.PlaceholderFormat
//
// It calls sq.Dollar.ReplacePlaceholders on the input sql string, and then replaces
// certain target placeholders with a type-cast version of it (e.g. $1 --> $1::uuid).
//
// If typeCast f does not contain any replacement pairs of strings, it will simply return the formatted string
// from a squirrel.Dollar's ReplacePlaceholders call. Otherwise, it will use a strings.Replacer to find a match and
// replace it with another value. In the context of type casting, a caller would configure typeCast with a slice of
// old-new pairs of strings such as `[]string{"$1,", "$1::uuid,"}`.
func (f typeCast) ReplacePlaceholders(s string) (string, error) {
	formatted, err := sq.Dollar.ReplacePlaceholders(s)
	if err != nil {
		return "", err
	}

	if f.oldnew == nil {
		return formatted, nil
	}

	return strings.NewReplacer(f.oldnew...).Replace(formatted), nil
}
