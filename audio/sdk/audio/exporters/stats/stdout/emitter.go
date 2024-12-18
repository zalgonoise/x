package stdout

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"log/slog"

	"github.com/zalgonoise/cfg"

	"github.com/zalgonoise/x/audio/fft"
	"github.com/zalgonoise/x/audio/sdk/audio"
	"github.com/zalgonoise/x/audio/sdk/audio/exporters"
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

func ToLogger(
	options []cfg.Option[*exporters.StatsConfig],
	logger *slog.Logger, metrics audio.ExporterMetrics, tracer trace.Tracer,
) (audio.Exporter, error) {
	return exporters.NewStatsExporter(
		emitter{logger: logger},
		options,
		logger, metrics, tracer,
	)
}
