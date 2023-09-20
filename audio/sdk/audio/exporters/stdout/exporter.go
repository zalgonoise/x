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

func (e logExporter) Shutdown(ctx context.Context) error {
	e.logger.InfoContext(ctx, "log exporter shutting down")
	e.cancel()

	errs := make([]error, 0, 2)
	if err := e.peaks.Shutdown(ctx); err != nil {
		errs = append(errs, err)
	}

	if err := e.spectrum.Shutdown(ctx); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (e logExporter) export(ctx context.Context) {
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

			e.logger.InfoContext(context.Background(), peaksMessage, slog.Float64("peak_value", v))
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

			e.logger.InfoContext(context.Background(), spectrumMessage, slog.Int("frequency", v[0].Freq))
		}
	}
}

func ToLogger(options ...cfg.Option[Config]) (audio.Exporter, error) {
	config := cfg.Set[Config](Config{}, options...)

	ctx, cancel := context.WithCancel(context.Background())

	exporter := &logExporter{
		config:   config,
		peaks:    newPeaksCollector(config),
		spectrum: newSpectrumCollector(config),
		logger:   newLogger(config.handler),
		cancel:   cancel,
	}

	go exporter.export(ctx)

	return exporter, nil
}

func newLogger(h slog.Handler) *slog.Logger {
	if h == nil {
		return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			AddSource: true,
		}))
	}

	return slog.New(h)
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
