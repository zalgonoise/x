package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/zalgonoise/x/cookies/internal/repository"
)

const (
	minAlloc = 64

	getCookiesQuery    = `SELECT userID, cookies, lastGift FROM cookies WHERE userID = ?`
	listCookiesQuery   = `SELECT userID, cookies FROM cookies`
	addCookiesQuery    = `UPDATE cookies SET cookies = ?, lastGift = ? WHERE userID = ?`
	insertCookiesQuery = `INSERT INTO cookies (userID, cookies, lastGift) VALUES (?, ?, ?)`
)

var ErrUnexpectedRowsAffected = errors.New("unexpected number of rows affected")

type Clock interface {
	Now() time.Time
}

type SQLite struct {
	db *sql.DB

	clock Clock
}

func (r *SQLite) GetCookies(ctx context.Context, user string) (int, time.Time, error) {
	var (
		userID   string
		current  int
		lastGift int
	)

	if err := r.db.QueryRowContext(ctx, getCookiesQuery, user).Scan(&userID, &current, &lastGift); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return 0, time.Time{}, repository.ErrNotFound
		default:
			return 0, time.Time{}, err
		}
	}

	return current, time.UnixMilli(int64(lastGift)), nil
}

func (r *SQLite) ListCookies(ctx context.Context) (map[string]int, error) {
	rows, err := r.db.QueryContext(ctx, listCookiesQuery)
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

func (r *SQLite) AddCookie(ctx context.Context, user string, n int) (int, error) {
	var (
		userID   string
		previous int
		lastGift int
	)

	if err := r.db.QueryRowContext(ctx, getCookiesQuery, user).Scan(&userID, &previous, &lastGift); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			// create user if not they don't exist
			res, err := r.db.ExecContext(ctx, insertCookiesQuery, user, n, int(r.clock.Now().UnixMilli()))
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

	res, err := r.db.ExecContext(ctx, addCookiesQuery, current, int(r.clock.Now().UnixMilli()), user)
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

func (r *SQLite) SwapCookies(ctx context.Context, from, to string, n int) (int, int, error) {
	var (
		requesterID       string
		requesterPrevious int
		requesterLastGift int
		targetID          string
		targetPrevious    int
		targetLastGift    int
	)

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, 0, err
	}

	defer tx.Rollback()

	if err := tx.QueryRowContext(ctx, getCookiesQuery, from).Scan(&requesterID, &requesterPrevious, &requesterLastGift); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return 0, 0, repository.ErrNotFound
		default:
			return 0, 0, err
		}
	}

	var cookiesAdded bool

	if err := tx.QueryRowContext(ctx, getCookiesQuery, to).Scan(&targetID, &targetPrevious, &targetLastGift); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			// create user if not they don't exist
			res, err := tx.ExecContext(ctx, insertCookiesQuery, targetID, n, int(r.clock.Now().UnixMilli()))
			if err != nil {
				return 0, 0, err
			}

			rowsAffected, err := res.RowsAffected()
			if err != nil {
				return 0, 0, err
			}

			if rowsAffected != 1 {
				return 0, 0, ErrUnexpectedRowsAffected
			}

			cookiesAdded = true
		default:
			return 0, 0, err
		}
	}

	requesterCurrent := requesterPrevious - n
	res, err := tx.ExecContext(ctx, addCookiesQuery, requesterCurrent, int(r.clock.Now().UnixMilli()), requesterID)
	if err != nil {
		return 0, 0, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, 0, err
	}

	if rowsAffected != 1 {
		return 0, 0, ErrUnexpectedRowsAffected
	}

	targetCurrent := n
	if !cookiesAdded {
		targetCurrent = targetPrevious + n
		res, err = tx.ExecContext(ctx, addCookiesQuery, targetCurrent, int(r.clock.Now().UnixMilli()), targetID)
		if err != nil {
			return 0, 0, err
		}

		rowsAffected, err = res.RowsAffected()
		if err != nil {
			return 0, 0, err
		}

		if rowsAffected != 1 {
			return 0, 0, ErrUnexpectedRowsAffected
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, err
	}

	return requesterCurrent, targetCurrent, nil
}

func (r *SQLite) EatCookie(ctx context.Context, user string) (int, error) {
	var userID string
	var previous int
	var lastGift int

	if err := r.db.QueryRowContext(ctx, getCookiesQuery, user).Scan(&userID, &previous, &lastGift); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return 0, repository.ErrNotFound
		default:
			return 0, err
		}
	}

	current := previous - 1

	res, err := r.db.ExecContext(ctx, addCookiesQuery, current, lastGift, user)
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
