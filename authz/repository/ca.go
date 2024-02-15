package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

const (
	queryGet = `
SELECT pub_key, cert FROM services
WHERE name = ?
`

	queryCreate = `
INSERT INTO services (name, pub_key, cert)
VALUES (?, ?, ?)
`

	queryDelete = `
DELETE FROM services WHERE name = ?
`
)

var (
	ErrNotFound       = errors.New("not found")
	ErrFailedDBWrite  = errors.New("failed to write entry")
	ErrFailedDBDelete = errors.New("failed to remove entry")
)

type CertificateAuthority struct {
	db *sql.DB
}

func NewCertificateAuthority(db *sql.DB) CertificateAuthority {
	return CertificateAuthority{db}
}

func (r CertificateAuthority) Get(ctx context.Context, service string) (pubKey []byte, cert []byte, err error) {
	if err = r.db.QueryRowContext(ctx, queryGet, service).Scan(&pubKey, &cert); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, ErrNotFound
		}

		return nil, nil, err
	}

	return pubKey, cert, nil
}

func (r CertificateAuthority) Create(ctx context.Context, service string, pubKey []byte, cert []byte) (err error) {
	res, err := r.db.ExecContext(ctx, queryCreate, service, pubKey, cert)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows != 1 {
		return fmt.Errorf("%w: %s", ErrFailedDBWrite, service)
	}

	return nil
}

func (r CertificateAuthority) Delete(ctx context.Context, service string) error {
	var pubKey, cert []byte

	if err := r.db.QueryRowContext(ctx, queryGet, service).Scan(&pubKey, &cert); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		return err
	}

	res, err := r.db.ExecContext(ctx, queryDelete, service)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows != 1 {
		return fmt.Errorf("%w: %s", ErrFailedDBDelete, service)
	}

	return nil
}
