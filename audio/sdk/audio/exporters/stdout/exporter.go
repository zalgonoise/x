package stdout

import (
	"cmp"
	"context"
	"errors"
	"log/slog"
	"slices"

	"github.com/zalgonoise/x/audio/fft"
	"github.com/zalgonoise/x/audio/sdk/audio"
	"github.com/zalgonoise/x/audio/sdk/audio/extractors"
	"github.com/zalgonoise/x/audio/sdk/audio/registries/batchreg"
	"github.com/zalgonoise/x/audio/wav/header"
	"github.com/zalgonoise/x/cfg"
)

const (
	peaksMessage    = "new peak registered"
	spectrumMessage = "new spectrum peak registered"
)

type logExporter struct {
	cfg Config

	peaks    audio.Collector[float64]
	spectrum audio.Collector[[]fft.FrequencyPower]

	logger *slog.Logger

	cancel context.CancelFunc
}

func (e logExporter) Export(h *header.Header, data []float64) error {
	errs := make([]error, 0, 2)

	if e.peaks != nil {
		if err := e.peaks.Collect(h, data); err != nil {
			errs = append(errs, err)
		}
	}

	if e.spectrum != nil {
		if err := e.spectrum.Collect(h, data); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (e logExporter) ForceFlush() error {
	if err := e.peaks.ForceFlush(); err != nil {
		return err
	}

	if err := e.spectrum.ForceFlush(); err != nil {
		return err
	}

	return nil
}

func (e logExporter) Shutdown(_ context.Context) error {
	e.cancel()

	return nil
}

func ToLogger(options ...cfg.Option[Config]) (audio.Exporter, error) {
	config := cfg.Set[Config](defaultConfig, options...)

	if err := Validate(config); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	exporter := logExporter{
		cfg:    config,
		cancel: cancel,
	}

	peaksCollector := audio.NoOpCollector[float64]()
	spectrumCollector := audio.NoOpCollector[[]fft.FrequencyPower]()

	if config.withPeaks {
		peaksCollector = audio.NewCollector[float64](
			extractors.MaxPeak(),
			batchreg.New[float64](nil, nil),
		)
	}

	exporter.peaks = peaksCollector

	if config.withSpectrum {
		spectrumCollector = audio.NewCollector[[]fft.FrequencyPower](
			extractors.MaxSpectrum(config.spectrumBlockSize),
			batchreg.New[[]fft.FrequencyPower](nil, nil),
		)
	}

	exporter.spectrum = spectrumCollector

	go func() {
		peaksValues := exporter.peaks.Load()
		spectrumValues := exporter.spectrum.Load()

		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-peaksValues:
				if ok {
					exporter.exportPeak(v)
				}
			case v, ok := <-spectrumValues:
				if ok {
					exporter.exportSpectrum(v)
				}
			}
		}
	}()

	return exporter, nil
}

func (e logExporter) exportPeak(value float64) {
	e.logger.InfoContext(context.Background(), peaksMessage, slog.Float64("peak_value", value))
}

func (e logExporter) exportSpectrum(value []fft.FrequencyPower) {
	slices.SortFunc(value, func(a, b fft.FrequencyPower) int {
		return cmp.Compare(a.Mag, b.Mag)
	})

	e.logger.InfoContext(context.Background(), spectrumMessage, slog.Int("frequency", value[0].Freq))
}
