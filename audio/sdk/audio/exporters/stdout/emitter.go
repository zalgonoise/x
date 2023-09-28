package stdout

import (
	"context"
	"log/slog"

	"github.com/zalgonoise/x/audio/fft"
	"github.com/zalgonoise/x/audio/sdk/audio"
	"github.com/zalgonoise/x/audio/sdk/audio/exporters"
	"github.com/zalgonoise/x/cfg"
)

const (
	peaksMessage    = "new peak registered"
	spectrumMessage = "new spectrum peak registered"
)

type emitter struct {
	logger *slog.Logger
}

func (e emitter) EmitPeaks(value float64) {
	e.logger.InfoContext(context.Background(), peaksMessage, slog.Float64("peak_value", value))
}

func (e emitter) EmitSpectrum(values []fft.FrequencyPower) {
	e.logger.InfoContext(context.Background(), spectrumMessage,
		slog.Int("frequency", values[0].Freq),
		slog.Float64("magnitude", values[0].Mag),
	)
}

func (e emitter) Shutdown(context.Context) error {
	return nil
}

func ToLogger(options ...cfg.Option[exporters.Config]) (audio.Exporter, error) {
	// re-use log handler from general exporter config
	config := cfg.Set[exporters.Config](exporters.DefaultConfig, options...)

	return exporters.PCM(
		emitter{logger: slog.New(config.LogHandler)},
		options...,
	)
}
