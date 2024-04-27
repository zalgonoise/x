package exporters

import (
	"context"
	"errors"
	"io"
	"sync/atomic"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/sdk/audio"
)

func NewDataExporter(writer io.Writer, options ...cfg.Option[DataConfig]) (audio.Exporter, error) {
	config := cfg.Set(defaultDataConfig(), options...)

	w, err := wav.New(config.sampleRate, config.bitDepth, config.numChannels, 1)
	if err != nil {
		return audio.NoOpExporter(), err
	}

	var maxSamples int64

	switch {
	case config.maxDuration != 0 && config.maxSamples == 0:
		maxSamples = int64(config.maxDuration.Seconds() * float64(config.sampleRate*uint32(config.numChannels)))
	case config.maxDuration == defaultDuration && config.maxSamples != 0:
		maxSamples = config.maxSamples
	default:
		maxSamples = numSeconds * int64(config.sampleRate*uint32(config.numChannels))
	}

	if config.extractor == nil {
		config.extractor = audio.NoOpExtractor[float64]()
	}

	if config.threshold == nil {
		config.threshold = audio.NoOpThreshold[float64]()
	}

	return &dataExporter{
		buffer:     w,
		writer:     writer,
		maxSamples: maxSamples,
		recording:  &atomic.Bool{},
		extractor:  config.extractor,
		threshold:  config.threshold,
	}, nil
}

type dataExporter struct {
	numSamples int64
	maxSamples int64
	recording  *atomic.Bool

	buffer    *wav.Wav
	writer    io.Writer
	extractor audio.Extractor[float64]
	threshold audio.Threshold[float64]
}

func (e *dataExporter) Export(header *wav.Header, data []float64) error {
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

func (e *dataExporter) ForceFlush() error {
	if e.writer == nil {
		return nil
	}

	_, err := io.Copy(e.writer, e.buffer)

	return err
}

func (e *dataExporter) Shutdown(context.Context) error {
	if e.writer == nil {
		return nil
	}

	return e.close()
}

func (e *dataExporter) close() error {
	if _, err := io.Copy(e.writer, e.buffer); err != nil {
		var closeErr error

		if closer, ok := e.writer.(io.Closer); ok {
			closeErr = closer.Close()
		}

		return errors.Join(err, closeErr)
	}

	e.buffer.Data.Reset()

	if closer, ok := e.writer.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			return err
		}
	}

	// if the writer is reusable (e.g. file writer that rotates filenames)
	// finds its `Reset() error` method to arm the writer again
	if resetter, ok := e.writer.(interface {
		Reset() error
	}); ok {
		if err := resetter.Reset(); err != nil {
			return err
		}
	}

	return nil
}
