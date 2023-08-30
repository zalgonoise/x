package stdout

import (
	"context"
	"log/slog"

	"github.com/zalgonoise/x/audio/fft"
	"github.com/zalgonoise/x/audio/sdk/audio"
	"github.com/zalgonoise/x/audio/wav/header"
	"github.com/zalgonoise/x/cfg"
)

type logExporter struct {
	cfg LoggerConfig

	peaks    audio.Collector
	spectrum audio.Collector

	logger *slog.Logger
}

func (e logExporter) Export(h *header.Header, data []float64) error {
	return nil
}

func (e logExporter) ForceFlush() error {
	return nil
}

func (e logExporter) Shutdown(ctx context.Context) error {
	return nil
}

func ToLogger(options ...cfg.Option[LoggerConfig]) (audio.Exporter, error) {
	config := cfg.Set[LoggerConfig](defaultConfig, options...)

	if err := Validate(config); err != nil {
		return nil, err
	}

	exporter := logExporter{
		cfg: config,
	}

	if config.withPeaks {
		exporter.peaks = audio.NewCollector[float64](
			audio.MaxPeak(),
			audio.NewRegistry[float64](), // replace with a batch registry or registry factory
		)
	}

	if config.withSpectrum {
		exporter.spectrum = audio.NewCollector[[]fft.FrequencyPower](
			audio.MaxSpectrum(config.spectrumBlockSize),
			audio.NewRegistry[[]fft.FrequencyPower](), // replace with a batch registry or registry factory
		)
	}

	return exporter, nil
}
