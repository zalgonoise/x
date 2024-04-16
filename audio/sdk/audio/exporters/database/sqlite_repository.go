package database

import (
	"database/sql"

	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/sdk/audio"
	"github.com/zalgonoise/x/errs"
)

const (
	errDomain = errs.Domain("x/audio/sdk/audio/exporters/database")

	ErrUnsupported = errs.Kind("unsupported")

	ErrHeader = errs.Entity("audio header")
)

var (
	ErrUnsupportedHeader = errs.WithDomain(errDomain, ErrUnsupported, ErrHeader)
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
)

type Repository struct {
	db *sql.DB
}

func (r Repository) Save(id string, header audio.Header, data []byte) error {
	_, ok := header.(*wav.Header)
	if !ok {
		return ErrUnsupportedHeader
	}

	// TODO: check and add header data
	// TODO: add chunk data if we have a header ID already

	return nil
}
