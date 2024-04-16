package exporters

import (
	"context"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/audio/encoding/wav/data"
	"github.com/zalgonoise/x/audio/sdk/audio"
)

type Repository interface {
	Save(id string, header audio.Header, data []byte) error
}

func NewSQLiteExporter(db Repository, options ...cfg.Option[SQLiteConfig]) (audio.Exporter, error) {
	// TODO:

	return nil, nil
}

type sqliteExporter struct {
	recording *atomic.Bool

	converter data.Converter
	repo      Repository
	extractor audio.Extractor[float64]
	threshold audio.Threshold[float64]
}

func (e *sqliteExporter) Export(header audio.Header, data []float64) error {
	// TODO: is the converter set once the first audio chunk comes in?

	// TODO: we may want to buffer a chunk of audio data before flushing it all at once into a row's bytes blob
	//   then, ForceFlush is also responsible for flushing the remaining bytes into the database if they exist.

	return e.repo.Save(uuid.New().String(), header, e.converter.Bytes(data))
}

func (e *sqliteExporter) ForceFlush() error {
	return nil
}

func (e *sqliteExporter) Shutdown(ctx context.Context) error {
	return nil
}
