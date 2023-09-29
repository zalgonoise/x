package exporters

import (
	"context"
	"errors"
	"log/slog"

	"github.com/zalgonoise/x/audio/errs"
	"github.com/zalgonoise/x/audio/fft"
	"github.com/zalgonoise/x/audio/sdk/audio"
	"github.com/zalgonoise/x/audio/sdk/audio/extractors"
	"github.com/zalgonoise/x/audio/sdk/audio/registries/batchreg"
	"github.com/zalgonoise/x/audio/sdk/audio/registries/unitreg"
	"github.com/zalgonoise/x/audio/wav/header"
	"github.com/zalgonoise/x/cfg"
)

const (
	errDomain = errs.Domain("x/audio/sdk/audio/exporters")

	ErrNil = errs.Kind("nil")

	ErrEmitter = errs.Entity("emitter")
)

var ErrNilEmitter = errs.New(errDomain, ErrNil, ErrEmitter)

type exporter struct {
	config Config

	peaks    audio.Collector[float64]
	spectrum audio.Collector[[]fft.FrequencyPower]

	emitter audio.Emitter

	logger *slog.Logger

	cancel context.CancelFunc
}

func (e exporter) Export(h audio.Header, data []float64) error {
	return errors.Join(
		e.peaks.Collect(h, data),
		e.spectrum.Collect(h, data),
	)
}

func (e exporter) ForceFlush() error {
	return errors.Join(
		e.peaks.ForceFlush(),
		e.spectrum.ForceFlush(),
	)
}

func (e exporter) Shutdown(ctx context.Context) error {
	e.logger.InfoContext(ctx, "exporter shutting down")
	e.cancel()

	return errors.Join(
		e.peaks.Shutdown(ctx),
		e.spectrum.Shutdown(ctx),
		e.emitter.Shutdown(ctx),
	)
}

func (e exporter) export(ctx context.Context) {
	peaksValues := e.peaks.Load()
	spectrumValues := e.spectrum.Load()

	for {
		select {
		case <-ctx.Done():
			return
		case v, ok := <-peaksValues:
			if !ok {
				return
			}

			e.emitter.EmitPeaks(v)
		case v, ok := <-spectrumValues:
			if !ok {
				return
			}

			e.emitter.EmitSpectrum(v)
		}
	}
}

func NewExporter(emitter audio.Emitter, options ...cfg.Option[Config]) (audio.Exporter, error) {
	if emitter == nil {
		return audio.NoOpExporter[*header.Header](), ErrNilEmitter
	}

	config := cfg.Set[Config](DefaultConfig, options...)

	ctx, cancel := context.WithCancel(context.Background())

	e := exporter{
		config:   config,
		peaks:    newPeaksCollector(config),
		spectrum: newSpectrumCollector(config),
		emitter:  emitter,
		logger:   slog.New(config.LogHandler),
		cancel:   cancel,
	}

	go e.export(ctx)

	return e, nil
}

func newPeaksCollector(config Config) audio.Collector[float64] {
	if !config.withPeaks {
		return audio.NoOpCollector[float64]()
	}

	if !config.batchedPeaks {
		return audio.NewCollector[float64](
			extractors.MaxPeak(),
			unitreg.New[float64](0),
		)
	}

	return audio.NewCollector[float64](
		extractors.MaxPeak(),
		batchreg.New[float64](config.batchedPeaksOptions...),
	)
}

func newSpectrumCollector(config Config) audio.Collector[[]fft.FrequencyPower] {
	if !config.withSpectrum {
		return audio.NoOpCollector[[]fft.FrequencyPower]()
	}

	if !config.batchedPeaks {
		return audio.NewCollector[[]fft.FrequencyPower](
			extractors.MaxSpectrum(config.spectrumBlockSize),
			unitreg.New[[]fft.FrequencyPower](0),
		)
	}

	return audio.NewCollector[[]fft.FrequencyPower](
		extractors.MaxSpectrum(config.spectrumBlockSize),
		batchreg.New[[]fft.FrequencyPower](config.batchedSpectrumOptions...),
	)
}
