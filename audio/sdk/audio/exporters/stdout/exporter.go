package stdout

import (
	"context"
	"errors"
	"log/slog"
	"os"

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
	return errors.Join(
		e.peaks.Collect(h, data),
		e.spectrum.Collect(h, data),
	)
}

func (e logExporter) ForceFlush() error {
	return errors.Join(
		e.peaks.ForceFlush(),
		e.spectrum.ForceFlush(),
	)
}

func (e logExporter) Shutdown(ctx context.Context) error {
	e.logger.InfoContext(ctx, "log exporter shutting down")
	e.cancel()

	return errors.Join(
		e.peaks.Shutdown(ctx),
		e.spectrum.Shutdown(ctx),
	)
}

func (e logExporter) export(ctx context.Context) {
	peaksValues := e.peaks.Load()
	spectrumValues := e.spectrum.Load()

	for {
		select {
		case <-ctx.Done():
			e.logger.InfoContext(ctx, "stopping exporter's routine")

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

			// the logger implementation will only print the data in the first item of the slice.
			//
			// it is the responsibility of the caller to configure an appropriate compactor for
			// custom behavior, e.g.:
			//
			//     exporter, err := stdout.ToLogger(
			//       stdout.WithBatchedSpectrum(
			//         // ...
			//         batchreg.WithCompactor[[]fft.FrequencyPower](compactors.MaxSpectra),
			//       ),
			//     )
			e.logger.InfoContext(context.Background(), spectrumMessage,
				slog.Int("frequency", v[0].Freq),
				slog.Float64("magnitude", v[0].Mag),
			)
		}
	}
}

func ToLogger(options ...cfg.Option[Config]) (audio.Exporter, error) {
	config := cfg.Set[Config](Config{}, options...)

	ctx, cancel := context.WithCancel(context.Background())

	exporter := logExporter{
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
		return slog.New(slog.NewTextHandler(os.Stderr, nil))
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
