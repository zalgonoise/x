package extractors

import (
	"cmp"
	"slices"

	"github.com/zalgonoise/x/audio/fft"
	"github.com/zalgonoise/x/audio/fft/window"
	"github.com/zalgonoise/x/audio/sdk/audio"
)

// MaxPeak returns a float64 Collector that calculates the maximum peak value in an audio signal
func MaxPeak() audio.Extractor[float64] {
	return audio.Extraction[float64](func(_ audio.Header, data []float64) (maximum float64) {
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
	return audio.Extraction[float64](func(_ audio.Header, data []float64) (average float64) {
		for i := range data {
			average += data[i]
		}

		return average / float64(len(data))
	})
}

// MaxSpectrum returns a []fft.FrequencyPower Collector that calculates the spectrum values in an audio signal
func MaxSpectrum(size int) audio.Extractor[[]fft.FrequencyPower] {
	if size < 8 {
		size = 64
	}

	return audio.Extraction[[]fft.FrequencyPower](func(h audio.Header, data []float64) []fft.FrequencyPower {
		sampleRate := h.GetSampleRate()
		if sampleRate == 0 {
			sampleRate = 44100
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
				return cmp.Compare(b.Mag, a.Mag)
			})

			maximum = append(maximum, spectrum[0])
		}

		return maximum
	})
}

// Spectrum returns a []fft.FrequencyPower Collector that calculates the full spectrum values in an audio signal
// with a given Compactor as reducer / filter
func Spectrum(size int, compactor audio.Compactor[[]fft.FrequencyPower]) audio.Extractor[[]fft.FrequencyPower] {
	if size < 8 {
		size = 64
	}

	return audio.Extraction[[]fft.FrequencyPower](func(h audio.Header, data []float64) []fft.FrequencyPower {
		sampleRate := h.GetSampleRate()
		if sampleRate == 0 {
			sampleRate = 44100
		}

		bs := fft.NearestBlock(size)
		windowBlock := window.New(window.Blackman, int(bs))

		spectra := make([][]fft.FrequencyPower, 0, len(data)/int(bs))

		for i := 0; i+int(bs) < len(data); i += int(bs) {
			spectrum := fft.Apply(
				sampleRate,
				data[i:i+int(bs)],
				windowBlock,
			)

			if len(spectrum) == 0 {
				continue
			}

			spectra = append(spectra, spectrum)
		}

		compact, err := compactor(spectra)
		if err != nil {
			return spectra[0]
		}

		return compact
	})
}
