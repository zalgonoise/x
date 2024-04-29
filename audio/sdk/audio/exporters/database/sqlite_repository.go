package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zalgonoise/x/audio/encoding/wav"
)

const (
	queryInsertHeader = `
INSERT INTO headers (
	uuid, timestamp,
	chunk_id, chunk_size, format,	subchunk_1_id, subchunk_1_size, 
	audio_format, num_channels, sample_rate, byte_rate, block_align, bits_per_sample
) VALUES (
	?, ?,
	?, ?, ?, ?, ?,
	?, ?, ?, ?, ?, ?
);`

	queryInsertChunk = `
	INSERT INTO chunks (
	uuid, header_id, timestamp,
	subchunk_2_id, subchunk_data
) VALUES (
	?, ?, ?,
	?, ?
);
`

	querySelectHeaderWhereID = `SELECT uuid FROM headers WHERE headers.uuid = ? LIMIT 1;`
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db}
}

func (r *Repository) Save(ctx context.Context, id string, header *wav.Header, data []byte) (string, error) {
	id, err := getOrCreateUUID(ctx, r.db, id, header)
	if err != nil {
		return "", err
	}

	if err := insertAudioData(ctx, r.db, id, data); err != nil {
		return "", err
	}

	return id, nil
}

func insertAudioData(ctx context.Context, db *sql.DB, id string, data []byte) error {
	res, err := db.ExecContext(ctx, queryInsertChunk, id, uuid.New().String(), time.Now(), data)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows != 1 {
		return fmt.Errorf("expected to affect 1 row, affected %d", rows)
	}

	return nil
}

func getOrCreateUUID(ctx context.Context, db *sql.DB, id string, header *wav.Header) (string, error) {
	var dbID string

	row := db.QueryRowContext(ctx, querySelectHeaderWhereID, id)
	if err := row.Scan(&dbID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return createHeader(ctx, db, header)
		}

		return "", err
	}

	return dbID, nil
}

func createHeader(ctx context.Context, db *sql.DB, header *wav.Header) (string, error) {
	id := uuid.New().String()

	res, err := db.ExecContext(ctx, queryInsertHeader,
		id, time.Now(), header.ChunkID[:],
		int64(header.ChunkSize), header.Format[:], header.Subchunk1ID[:], int64(header.Subchunk1Size),
		int64(header.AudioFormat), int64(header.NumChannels), int64(header.SampleRate),
		int64(header.SampleRate), int64(header.ByteRate), int64(header.BlockAlign), int64(header.BitsPerSample),
	)
	if err != nil {
		return "", err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return "", err
	}
	if rows != 1 {
		return "", fmt.Errorf("expected to affect 1 row, affected %d", rows)
	}

	return id, nil
}
