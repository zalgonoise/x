package processors

import (
	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/encoding/wav/header"
	"github.com/zalgonoise/x/audio/sdk/audio"
)

func PCM(e ...audio.Exporter) audio.Processor {
	if len(e) == 0 {
		return audio.NoOpProcessor()
	}

	exporter := audio.MultiExporter(e...)

	return audio.NewProcessor(
		audio.NewStreamExporter(
			wav.NewStream(nil, func(h *header.Header, data []float64) error {
				return exporter.Export(h, data)
			}),
			exporter,
		),
	)
}
