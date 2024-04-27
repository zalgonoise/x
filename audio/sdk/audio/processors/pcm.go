package processors

import (
	"github.com/zalgonoise/cfg"

	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/sdk/audio"
)

func PCM(exporters []audio.Exporter, opts ...cfg.Option[wav.Config]) audio.Processor {
	if len(exporters) == 0 {
		return audio.NoOpProcessor()
	}

	exporter := audio.MultiExporter(exporters...)

	return audio.NewProcessor(
		audio.NewStreamExporter(
			wav.NewStream(exporter.Export, opts...),
			exporter,
		),
	)
}
