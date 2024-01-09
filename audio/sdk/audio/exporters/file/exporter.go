package file

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/sdk/audio"
)

const defaultSlug = "audio-capture"

type exporter struct {
	sampleRate  uint32
	bitDepth    uint16
	numChannels uint16

	dir  string
	name string
	idx  int

	buffer *wav.Wav
	f      io.WriteCloser
}

func (e *exporter) Export(_ audio.Header, data []float64) error {
	e.buffer.Data.ParseFloat(data)

	return nil
}

func (e *exporter) ForceFlush() error {
	if e.f == nil {
		if err := e.open(); err != nil {
			return err
		}
	}

	_, err := io.Copy(e.f, e.buffer)

	return err
}

func (e *exporter) Shutdown(_ context.Context) error {
	if e.f == nil {
		if err := e.open(); err != nil {
			return err
		}
	}

	_, err := io.Copy(e.f, e.buffer)

	e.buffer.Data.Reset()

	if err != nil {
		defer e.f.Close()

		return err
	}

	if err = e.f.Close(); err != nil {
		return err
	}

	e.f = nil

	return nil
}

func ToFile(opts ...cfg.Option[Config]) (audio.Exporter, error) {
	config := cfg.Set(defaultConfig(), opts...)

	w, err := wav.New(config.sampleRate, config.bitDepth, config.numChannels, 1)

	if err != nil {
		return audio.NoOpExporter(), err
	}

	return &exporter{
		dir:    config.outputDir,
		name:   config.filenamePrefix,
		buffer: w,
	}, nil
}

func (e *exporter) open() error {
	if e.dir == "" {
		e.dir = os.TempDir()
	}

	if e.name == "" {
		e.name = defaultSlug
	}

	target := fmt.Sprintf("%s/%s_%04d", e.dir, e.name, e.idx)

	f, err := os.Open(target)

	if errors.Is(err, fs.ErrNotExist) {
		f, err = os.Create(target)
	}

	if err != nil {
		return err
	}

	e.idx++
	e.f = f

	return nil
}
