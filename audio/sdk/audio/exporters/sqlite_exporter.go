package exporters

import (
	"context"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/encoding/wav/data/conv"
	"github.com/zalgonoise/x/audio/sdk/audio"
	"github.com/zalgonoise/x/audio/sdk/audio/exporters/data"
	"github.com/zalgonoise/x/errs"
)

const (
	errDomain = errs.Domain("x/audio/sdk/audio/exporters")

	ErrUnsupported = errs.Kind("unsupported")

	ErrFormat = errs.Entity("audio format")
)

var (
	ErrUnsupportedFormat = errs.WithDomain(errDomain, ErrUnsupported, ErrFormat)
)

type Repository interface {
	Save(ctx context.Context, id string, header *wav.Header, data []byte) (string, error)
}

type Converter interface {
	Bytes(buf []float64) []byte
}

type Flusher interface {
	Write(id string, h *wav.Header, data []byte) (int, error)
	ForceFlush(id string, h *wav.Header) error
}

const (
	bitDepth8  uint16 = 8
	bitDepth16 uint16 = 16
	bitDepth24 uint16 = 24
	bitDepth32 uint16 = 32
	bitDepth64 uint16 = 64
)

func NewSQLiteExporter(db Repository, options ...cfg.Option[SQLiteConfig]) (audio.Exporter, error) {
	config := cfg.Set(defaultConfig(), options...)

	e := &sqliteExporter{
		recordID: &atomic.Pointer[string]{},
		repo:     db,
	}

	e.flusher = data.NewGZipFlusher(
		config.dur,
		func(id string, h *wav.Header, data []byte) error {
			id, err := e.repo.Save(context.Background(), id, h, data)
			if err != nil {
				return err
			}

			e.recordID.Store(&id)

			return nil
		})

	return e, nil
}

type sqliteExporter struct {
	recordID *atomic.Pointer[string]

	flusher   Flusher
	converter Converter
	repo      Repository
}

func (e *sqliteExporter) Export(ctx context.Context, header *wav.Header, data []float64) error {
	id := e.recordID.Load()
	recID, ok := wav.GetID(ctx)
	if !ok {
		recID = uuid.New().String()
	}

	_ = e.recordID.CompareAndSwap(id, &recID)

	if e.converter == nil {
		var err error

		e.converter, err = converterFrom(header)
		if err != nil {
			return err
		}
	}

	_, err := e.flusher.Write(recID, header, e.converter.Bytes(data))
	if err != nil {
		return err
	}

	return nil
}

func converterFrom(h *wav.Header) (Converter, error) {
	switch h.AudioFormat {
	case wav.UnsetFormat, wav.PCMFormat:
		switch h.BitsPerSample {
		case bitDepth8:
			return conv.PCM8Bit{}, nil
		case bitDepth16:
			return conv.PCM16Bit{}, nil
		case bitDepth24:
			return conv.PCM24Bit{}, nil
		case bitDepth32:
			return conv.PCM32Bit{}, nil
		default:
			return nil, ErrUnsupportedFormat
		}
	case wav.FloatFormat:
		switch h.BitsPerSample {
		case bitDepth32:
			return conv.Float32{}, nil
		case bitDepth64:
			return conv.Float64{}, nil
		default:
			return nil, ErrUnsupportedFormat
		}
	}

	return nil, ErrUnsupportedFormat
}

func (e *sqliteExporter) ForceFlush() error {
	id := e.recordID.Load()
	if id == nil {
		return nil
	}

	return e.flusher.ForceFlush(*id, nil)
}

func (e *sqliteExporter) Shutdown(ctx context.Context) error {
	return nil
}
