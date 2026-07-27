package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/zalgonoise/x/politebot/internal/repository"
)

const (
	minAlloc = 64

	getAngyPointsQuery    = `SELECT user_id, angy_points, last_punishment FROM angy WHERE userID = ?`
	listAngyPointsQuery   = `SELECT user_id, angy_points FROM angy`
	addAngyPointsQuery    = `UPDATE angy SET angy_points = ?, last_punishment = ? WHERE user_id = ?`
	insertAngyPointsQuery = `INSERT INTO angy (user_id, angy_points, last_punishment) VALUES (?, ?, ?)`
)

var ErrUnexpectedRowsAffected = errors.New("unexpected number of rows affected")

type Clock interface {
	Now() time.Time
}

type SQLite struct {
	db *sql.DB

	clock Clock
}

func (r *SQLite) GetAngyPoints(ctx context.Context, user string) (int, time.Time, error) {
	var (
		userID         string
		current        int
		lastPunishment int
	)

	if err := r.db.QueryRowContext(ctx, getAngyPointsQuery, user).Scan(&userID, &current, &lastPunishment); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return 0, time.Time{}, repository.ErrNotFound
		default:
			return 0, time.Time{}, err
		}
	}

	return current, time.UnixMilli(int64(lastPunishment)), nil
}

func (r *SQLite) ListAngyPoints(ctx context.Context) (map[string]int, error) {
	rows, err := r.db.QueryContext(ctx, listAngyPointsQuery)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	res := make(map[string]int, minAlloc)

	for rows.Next() {
		var (
			userID  string
			current int
		)

		if err := rows.Scan(&userID, &current); err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				return nil, repository.ErrNotFound
			default:
				return nil, err
			}
		}

		res[userID] = current
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (r *SQLite) AddAngyPoints(ctx context.Context, user string, n int) (int, error) {
	var (
		userID         string
		previous       int
		lastPunishment int
	)

	if err := r.db.QueryRowContext(ctx, getAngyPointsQuery, user).Scan(&userID, &previous, &lastPunishment); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			// create user if not they don't exist
			res, err := r.db.ExecContext(ctx, insertAngyPointsQuery, user, n, int(r.clock.Now().UnixMilli()))
			if err != nil {
				return 0, err
			}

			rowsAffected, err := res.RowsAffected()
			if err != nil {
				return 0, err
			}

			if rowsAffected != 1 {
				return 0, ErrUnexpectedRowsAffected
			}

			return n, nil
		default:
			return 0, err
		}
	}

	current := previous + n

	res, err := r.db.ExecContext(ctx, addAngyPointsQuery, current, int(r.clock.Now().UnixMilli()), user)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	if rowsAffected != 1 {
		return 0, ErrUnexpectedRowsAffected
	}

	return current, nil
}

func NewSQLite(db *sql.DB, clock Clock) *SQLite {
	return &SQLite{db: db, clock: clock}
}
