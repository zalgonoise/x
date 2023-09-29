package processors

import (
	"github.com/zalgonoise/x/audio/sdk/audio"
	"github.com/zalgonoise/x/audio/sdk/audio/exporters"
	"github.com/zalgonoise/x/audio/wav"
	"github.com/zalgonoise/x/audio/wav/header"
)

func PCM(e ...audio.Exporter) audio.Processor {
	if len(e) == 0 {
		return audio.NoOpProcessor()
	}

	exporter := exporters.Multi(e...)

	return NewProcessor(
		wav.NewStream(nil, func(h *header.Header, data []float64) error {
			return exporter.Export(h, data)
		}),
		exporter,
	)
}
