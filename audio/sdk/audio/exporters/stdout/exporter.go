package stdout

import (
	"cmp"
	"context"
	"errors"
	"log/slog"
	"os"
	"slices"

	"github.com/zalgonoise/x/audio/fft"
	"github.com/zalgonoise/x/audio/sdk/audio"
	"github.com/zalgonoise/x/audio/sdk/audio/extractors"
	"github.com/zalgonoise/x/audio/sdk/audio/registries/batchreg"
	"github.com/zalgonoise/x/audio/sdk/audio/registries/unitreg"
	"github.com/zalgonoise/x/audio/wav/header"
	"github.com/zalgonoise/x/cfg"
)

const (
	peaksMessage    = "new peak registered"
	spectrumMessage = "new spectrum peak registered"
)

type logExporter struct {
	config Config

	peaks    audio.Collector[float64]
	spectrum audio.Collector[[]fft.FrequencyPower]

	logger *slog.Logger

	cancel context.CancelFunc
}

func (e logExporter) Export(h *header.Header, data []float64) error {
	errs := make([]error, 0, 2)

	if err := e.peaks.Collect(h, data); err != nil {
		errs = append(errs, err)
	}

	if err := e.spectrum.Collect(h, data); err != nil {
		errs = append(errs, err)
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
		return audio.NoOpExporter(), err
	}

	switch {
	case config.logger == nil && config.handler == nil:
		config.logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			AddSource: true,
		}))
	case config.logger == nil && config.handler != nil:
		config.logger = slog.New(config.handler)
	}

	ctx, cancel := context.WithCancel(context.Background())

	exporter := &logExporter{
		logger: config.logger,
		config: config,
		cancel: cancel,
	}

	var (
		peaksCollector    = audio.NoOpCollector[float64]()
		spectrumCollector = audio.NoOpCollector[[]fft.FrequencyPower]()
	)

	if config.withPeaks {
		var reg audio.Registerer[float64]

		if config.batchedPeaks {
			reg = batchreg.New[float64](config.batchedPeaksOptions...)
		} else {
			reg = unitreg.New[float64](0)
		}

		peaksCollector = audio.NewCollector[float64](extractors.MaxPeak(), reg)
	}

	exporter.peaks = peaksCollector

	if config.withSpectrum {
		var reg audio.Registerer[[]fft.FrequencyPower]

		if config.batchedSpectrum {
			reg = batchreg.New[[]fft.FrequencyPower](config.batchedSpectrumOptions...)
		} else {
			reg = unitreg.New[[]fft.FrequencyPower](0)
		}

		spectrumCollector = audio.NewCollector[[]fft.FrequencyPower](extractors.MaxSpectrum(config.spectrumBlockSize), reg)
	}

	exporter.spectrum = spectrumCollector

	go func() {
		peaksValues := peaksCollector.Load()
		spectrumValues := spectrumCollector.Load()

		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-peaksValues:
				if !ok {
					return
				}

				exporter.logger.InfoContext(context.Background(), peaksMessage, slog.Float64("peak_value", v))
			case v, ok := <-spectrumValues:
				if !ok {
					return
				}

				if len(v) == 0 {
					continue
				}

				slices.SortFunc(v, func(a, b fft.FrequencyPower) int {
					return cmp.Compare(a.Mag, b.Mag)
				})

				exporter.logger.InfoContext(context.Background(), spectrumMessage, slog.Int("frequency", v[0].Freq))
			}
		}
	}()

	return exporter, nil
}
