package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/zalgonoise/x/randoom/items"
)

var (
	ErrUnimplemented  = errors.New("unimplemented")
	ErrNoDataInserted = errors.New("no data was inserted")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Load(ctx context.Context, list items.List) error {
	labelID, err := r.storeLabel(ctx, list.Label)
	if err != nil {
		return err
	}

	if err = r.storeLabelItems(ctx, labelID, list.Items); err != nil {
		return err
	}

	return ErrUnimplemented
}

func (r *Repository) storeLabel(ctx context.Context, label string) (id int64, err error) {
	rows, err := r.db.QueryContext(ctx, `INSERT INTO labels (label) VALUES (?) RETURNING id`, label)
	if err != nil {
		return -1, err
	}

	defer rows.Close()

	var labelID int64
	for rows.Next() {
		if err := rows.Scan(&labelID); err != nil {
			return -1, err
		}

		break
	}

	if err = rows.Close(); err != nil {
		return -1, err
	}

	if err = rows.Err(); err != nil {
		return -1, err
	}

	return labelID, nil
}

func (r *Repository) storeLabelItems(ctx context.Context, labelID int64, items []items.Item) error {
	var insertString = `INSERT INTO label_items (label_id, content, count, ratio) VALUES `

	args := make([]any, 0, len(items)*4)
	sb := &strings.Builder{}

	sb.WriteString(insertString)

	for i := range items {
		sb.WriteString(fmt.Sprintf(`(%d, %q, %d, %f)`, labelID, items[i].Content, items[i].Count, items[i].Ratio))
		args = append(args, labelID, items[i].Content, items[i].Count, items[i].Ratio)

		if i != len(items)-1 {
			sb.WriteString(", ")
		}
	}

	res, err := r.db.ExecContext(ctx, sb.String(), args...)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return ErrNoDataInserted
	}

	return nil
}

// TODO: add a Register method for manual, non-randomized increments

func (r *Repository) GetRandomItem(ctx context.Context, label string) (*items.Item, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT li.content, li.count, li.ratio FROM label_items AS li 
    WHERE li.label_id = (
    	SELECT id FROM labels WHERE label = ?
    ) ORDER BY (li.count * li.ratio) DESC`, label)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var item = &items.Item{}
	for rows.Next() {
		if err := rows.Scan(&item.Content, &item.Count, &item.Ratio); err != nil {
			return nil, err
		}

		break
	}

	if err = rows.Close(); err != nil {
		return nil, err
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return item, nil
}

func (r *Repository) CreatePlaylist(ctx context.Context, label string, size int) (int64, []*items.Item, error) {
	labelID, err := r.getLabelID(ctx, label)
	if err != nil {
		return -1, nil, err
	}

	playlistID, err := r.createPlaylist(ctx, labelID)
	if err != nil {
		return -1, nil, err
	}

	i, err := r.populatePlaylist(ctx, playlistID, labelID, size)
	if err != nil {
		return -1, nil, err
	}

	return playlistID, i, nil
}

func (r *Repository) createPlaylist(ctx context.Context, labelID int64) (int64, error) {
	rows, err := r.db.QueryContext(ctx, `INSERT INTO playlists (label_id) VALUES (?) RETURNING id`, labelID)
	if err != nil {
		return -1, err
	}

	defer rows.Close()

	var playlistID int64
	for rows.Next() {
		if err := rows.Scan(&playlistID); err != nil {
			return -1, err
		}

		break
	}

	if err = rows.Close(); err != nil {
		return -1, err
	}

	if err = rows.Err(); err != nil {
		return -1, err
	}

	return playlistID, nil
}

func (r *Repository) getLabelID(ctx context.Context, label string) (int64, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id FROM labels WHERE label = ?`, label)
	if err != nil {
		return -1, err
	}

	defer rows.Close()

	var id int64

	for rows.Next() {
		if err := rows.Scan(&id); err != nil {
			return -1, err
		}

		break
	}

	if err = rows.Close(); err != nil {
		return -1, err
	}

	if err = rows.Err(); err != nil {
		return -1, err
	}

	return id, nil
}

func (r *Repository) populatePlaylist(ctx context.Context, playlistID, labelID int64, size int) ([]*items.Item, error) {
	rows, err := r.db.QueryContext(ctx, `INSERT INTO playlist_items (playlist_id, item_id) VALUES (?, (
SELECT li.id, li.content, li.count, li.ratio FROM label_items AS li 
    WHERE li.label_id = ?
     ORDER BY (li.count * li.ratio) DESC)
))`, playlistID, labelID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var i = make([]*items.Item, 0, size)
	for rows.Next() {
		item := &items.Item{}

		if err := rows.Scan(&item.ID, &item.Content, &item.Count, &item.Ratio); err != nil {
			return nil, err
		}

		i = append(i, item)

		if len(i) >= size {
			break
		}
	}

	if err = rows.Close(); err != nil {
		return nil, err
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return i, nil
}

func (r *Repository) DeletePlaylist(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM playlist_items WHERE playlist_id = ?`, id)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	res, err = r.db.ExecContext(ctx, `DELETE FROM playlists WHERE id = ?`, id)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()

	return err
}
