package exporters

import (
	"github.com/zalgonoise/cfg"

	"github.com/zalgonoise/x/audio/fft"
	"github.com/zalgonoise/x/audio/sdk/audio"
	"github.com/zalgonoise/x/audio/sdk/audio/extractors"
	"github.com/zalgonoise/x/audio/sdk/audio/registries/batchreg"
	"github.com/zalgonoise/x/audio/sdk/audio/registries/unitreg"
)

func NewStatsExporter(emitter audio.Emitter, options ...cfg.Option[*StatsConfig]) (audio.Exporter, error) {
	config := cfg.Set[*StatsConfig](DefaultStatsConfig(), options...)

	e, err := audio.NewExporter(emitter, newPeaksCollector(config), newSpectrumCollector(config))
	if err != nil {
		return audio.NoOpExporter(), err
	}

	if config.LogHandler != nil {
		e = audio.ExporterWithLogs(e, config.LogHandler)
	}

	return e, nil
}

func newPeaksCollector(config *StatsConfig) audio.Collector[float64] {
	if config == nil || !config.withPeaks {
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

func newSpectrumCollector(config *StatsConfig) audio.Collector[[]fft.FrequencyPower] {
	if config == nil || !config.withSpectrum {
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
