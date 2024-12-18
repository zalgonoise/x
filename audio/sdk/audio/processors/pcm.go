package processors

import (
	"github.com/zalgonoise/cfg"
	"go.opentelemetry.io/otel/trace"
	"log/slog"

	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/sdk/audio"
)

func PCM(
	exporters []audio.Exporter, streamOptions []cfg.Option[wav.Config],
	logger *slog.Logger, metrics audio.ProcessorMetrics, tracer trace.Tracer,
) audio.Processor {
	if len(exporters) == 0 {
		return audio.NoOpProcessor()
	}

	exporter := audio.MultiExporter(exporters...)

	return audio.NewProcessor(
		audio.NewStreamExporter(
			wav.NewStreamContext(exporter.Export, streamOptions...),
			exporter,
		),
		logger, metrics, tracer,
	)
}
