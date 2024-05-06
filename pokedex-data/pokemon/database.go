package pokemon

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	_ "github.com/jackc/pgx/v5/stdlib" // Postgres driver
)

const insertQuery = `INSERT INTO items (id, image_source, name) VALUES `

func buildInsert(summaries []Summary) (string, []any, error) {
	b := sq.Insert("items").Columns("id", "image_source", "name")

	for i := range summaries {
		b.Values(summaries[i].ID, summaries[i].Sprite, summaries[i].Name)
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

func OpenPostgres(uri string, maxConns int) (*sql.DB, error) {
	db, err := sql.Open("pgx", uri)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxConns)
	db.SetMaxIdleConns(maxConns)

	return db, nil
}
