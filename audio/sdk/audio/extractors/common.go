package extractors

import (
	"slices"

	"github.com/zalgonoise/x/audio/fft"
	"github.com/zalgonoise/x/audio/fft/window"
	"github.com/zalgonoise/x/audio/sdk/audio"
	"github.com/zalgonoise/x/audio/wav/header"
)

// MaxPeak returns a float64 Collector that calculates the maximum peak value in an audio signal
func MaxPeak() audio.Extractor[float64] {
	return audio.Extraction[float64](func(_ *header.Header, data []float64) (maximum float64) {
		for i := range data {
			if data[i] > maximum {
				maximum = data[i]
			}
		}

		return maximum
	})
}

// AveragePeak returns a float64 Collector that calculates the average peak value in an audio signal
func AveragePeak() audio.Extractor[float64] {
	return audio.Extraction[float64](func(_ *header.Header, data []float64) (average float64) {
		for i := range data {
			average += data[i]
		}

		return average / float64(len(data))
	})
}

// MaxSpectrum returns a []fft.FrequencyPower Collector that calculates the maximum spectrum values in an audio signal
func MaxSpectrum(size int) audio.Extractor[[]fft.FrequencyPower] {
	if size < 8 {
		size = 64
	}

	sampleRate := 44100

	return audio.Extraction[[]fft.FrequencyPower](func(h *header.Header, data []float64) []fft.FrequencyPower {
		if h != nil {
			sampleRate = int(h.SampleRate)
		}

		bs := fft.NearestBlock(size)
		windowBlock := window.New(window.Blackman, int(bs))

		maximum := make([]fft.FrequencyPower, 0, len(data)/int(bs))

		for i := 0; i+int(bs) < len(data); i += int(bs) {
			spectrum := fft.Apply(
				sampleRate,
				data[i:i+int(bs)],
				windowBlock,
			)

			slices.SortFunc(spectrum, func(a, b fft.FrequencyPower) int {
				switch {
				case a.Mag > b.Mag:
					return -1
				case a.Mag < b.Mag:
					return 1
				default:
					return 0
				}
			})

			maximum = append(maximum, spectrum[0])
		}

		return maximum
	})
}
