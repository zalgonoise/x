package file

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"sync/atomic"

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

	numSamples int64
	maxSamples int64
	recording  *atomic.Bool

	buffer    *wav.Wav
	f         io.WriteCloser
	extractor audio.Extractor[float64]
	threshold func(float64) bool
}

func (e *exporter) Export(header audio.Header, data []float64) error {
	value := e.extractor.Extract(header, data)

	if e.threshold(value) {
		e.recording.Store(true)
		e.numSamples = 0
	}

	if e.recording.Load() {
		e.buffer.Data.ParseFloat(data)
		e.numSamples += int64(len(data))
	}

	if e.numSamples > e.maxSamples {
		return e.close()
	}

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

	return e.close()
}

func ToFile(opts ...cfg.Option[Config]) (audio.Exporter, error) {
	config := cfg.Set(defaultConfig(), opts...)

	w, err := wav.New(config.sampleRate, config.bitDepth, config.numChannels, 1)

	if err != nil {
		return audio.NoOpExporter(), err
	}

	var maxSamples int64

	switch {
	case config.maxDuration != 0 && config.maxSamples == 0:
		maxSamples = int64(config.maxDuration.Seconds() * float64(wav.ByteRate(config.sampleRate, config.bitDepth, config.numChannels)))
	case config.maxDuration == defaultDuration && config.maxSamples != 0:
		maxSamples = config.maxSamples
	default:
		maxSamples = numSeconds * int64(wav.ByteRate(config.sampleRate, config.bitDepth, config.numChannels))
	}

	return &exporter{
		dir:        config.outputDir,
		name:       config.filenamePrefix,
		buffer:     w,
		maxSamples: maxSamples,
		recording:  &atomic.Bool{},
		extractor:  config.extractor,
		threshold:  config.threshold,
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

func (e *exporter) close() error {
	if _, err := io.Copy(e.f, e.buffer); err != nil {
		return errors.Join(err, e.f.Close())
	}

	e.buffer.Data.Reset()

	err := e.f.Close()

	e.f = nil

	return err
}
