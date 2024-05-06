package pokemon

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	sq "github.com/Masterminds/squirrel"
	"github.com/zalgonoise/x/pokedex-data/database"

	_ "github.com/jackc/pgx/v5/stdlib" // Postgres driver
)

func buildInsert(summaries []Summary) (string, []any, error) {
	b := sq.Insert("items").Columns("id", "image_source", "name")

	for i := range summaries {
		b = b.Values(summaries[i].ID, summaries[i].Sprite, summaries[i].Name)
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

func OpenPostgres(ctx context.Context, uri string, maxConns int, logger *slog.Logger) (*sql.DB, error) {
	db, err := sql.Open("pgx", uri)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxConns)
	db.SetMaxIdleConns(maxConns)

	if err = database.Migrate(ctx, db, logger); err != nil {
		return nil, err
	}

	return db, nil
}
