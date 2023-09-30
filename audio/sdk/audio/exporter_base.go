package audio

import (
	"context"
	"errors"

	"github.com/zalgonoise/x/audio/errs"
	"github.com/zalgonoise/x/audio/fft"
)

const (
	errDomain = errs.Domain("x/audio/sdk/audio")

	ErrNil = errs.Kind("nil")

	ErrEmitter = errs.Entity("emitter")
)

var ErrNilEmitter = errs.New(errDomain, ErrNil, ErrEmitter)

type exporter struct {
	peaks    Collector[float64]
	spectrum Collector[[]fft.FrequencyPower]

	emitter Emitter

	cancel context.CancelFunc
}

func (e exporter) Export(h Header, data []float64) error {
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

func NewExporter(
	emitter Emitter, peaks Collector[float64], spectrum Collector[[]fft.FrequencyPower],
) (Exporter, error) {
	if emitter == nil {
		return NoOpExporter(), ErrNilEmitter
	}

	if peaks == nil {
		peaks = NoOpCollector[float64]()
	}

	if spectrum == nil {
		spectrum = NoOpCollector[[]fft.FrequencyPower]()
	}

	ctx, cancel := context.WithCancel(context.Background())

	e := exporter{
		peaks:    peaks,
		spectrum: spectrum,
		emitter:  emitter,
		cancel:   cancel,
	}

	go e.export(ctx)

	return e, nil
}
