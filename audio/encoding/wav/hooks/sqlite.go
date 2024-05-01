package hooks

import (
	"context"
	"time"

	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/sdk/audio/exporters/data"
)

type Repository interface {
	Save(ctx context.Context, id string, header *wav.Header, data []byte) (string, error)
}

type SQLiteHook struct {
	flusher *data.Flusher

	repo Repository
}

func (s *SQLiteHook) Save(ctx context.Context, h *wav.Header, data []byte) error {
	if _, err := s.flusher.Write(wav.GetOrCreateID(ctx), h, data); err != nil {
		return err
	}

	return nil
}

func NewSQLiteHook(db Repository, dur time.Duration) *SQLiteHook {
	s := &SQLiteHook{
		repo: db,
	}

	s.flusher = data.NewFlusher(dur, func(id string, h *wav.Header, data []byte) error {
		_, err := s.repo.Save(context.Background(), id, h, data)
		if err != nil {
			return err
		}

		return nil
	})

	return s
}
