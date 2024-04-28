package extractors

import (
	"cmp"
	"context"
	"slices"

	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/fft"
	"github.com/zalgonoise/x/audio/fft/window"
	"github.com/zalgonoise/x/audio/sdk/audio"
)

const (
	minBlockSize     = 8
	defaultBlockSize = 64
)

// MaxPeak returns a float64 Collector that calculates the maximum peak value in an audio signal.
func MaxPeak() audio.Extractor[float64] {
	return audio.Extraction[float64](func(ctx context.Context, _ *wav.Header, data []float64) (maximum float64) {
		for i := range data {
			if data[i] > maximum {
				maximum = data[i]
			}
		}

		return maximum
	})
}

// MaxAbsPeak returns a float64 Collector that calculates the maximum absolute peak value in an audio signal, that is,
// where its negative values are normalized as positive ones, to find the peaks in both positive and negative axis of
// the wave.
//
// The returned value is the original data point in the signal, so if its peak is a negative value, a negative value is
// returned.
func MaxAbsPeak() audio.Extractor[float64] {
	return audio.Extraction[float64](func(ctx context.Context, _ *wav.Header, data []float64) (maximum float64) {
		var maxIdx int

		for i := range data {
			value := data[i]

			if value < 0.0 {
				value = -value
			}

			if value > maximum {
				maximum = value
				maxIdx = i
			}
		}

		return data[maxIdx]
	})
}

// AveragePeak returns a float64 Collector that calculates the average peak value in an audio signal.
func AveragePeak() audio.Extractor[float64] {
	return audio.Extraction[float64](func(ctx context.Context, _ *wav.Header, data []float64) (average float64) {
		for i := range data {
			average += data[i]
		}

		return average / float64(len(data))
	})
}

// MaxSpectrum returns a []fft.FrequencyPower Collector that calculates the spectrum values in an audio signal.
func MaxSpectrum(size int) audio.Extractor[[]fft.FrequencyPower] {
	if size < minBlockSize {
		size = defaultBlockSize
	}

	return audio.Extraction[[]fft.FrequencyPower](
		func(ctx context.Context, h *wav.Header, data []float64) []fft.FrequencyPower {
			if h.SampleRate == 0 {
				h.SampleRate = 44100
			}

			bs := fft.NearestBlock(size)
			windowBlock := window.New(window.Blackman, bs)

			maximum := make([]fft.FrequencyPower, 0, len(data)/bs)

			for i := 0; i+bs < len(data); i += bs {
				spectrum := fft.Apply(
					int(h.SampleRate),
					data[i:i+bs],
					windowBlock,
				)

				slices.SortFunc(spectrum, func(a, b fft.FrequencyPower) int {
					return cmp.Compare(b.Mag, a.Mag)
				})

				maximum = append(maximum, spectrum[0])
			}

			return maximum
		},
	)
}

// Spectrum returns a []fft.FrequencyPower Collector that calculates the full spectrum values in an audio signal
// with a given Compactor as reducer / filter.
func Spectrum(size int, compactor audio.Compactor[[]fft.FrequencyPower]) audio.Extractor[[]fft.FrequencyPower] {
	if size < minBlockSize {
		size = defaultBlockSize
	}

	return audio.Extraction[[]fft.FrequencyPower](
		func(ctx context.Context, h *wav.Header, data []float64) []fft.FrequencyPower {
			if h.SampleRate == 0 {
				h.SampleRate = 44100
			}

			bs := fft.NearestBlock(size)
			windowBlock := window.New(window.Blackman, bs)

			spectra := make([][]fft.FrequencyPower, 0, len(data)/bs)

			for i := 0; i+bs < len(data); i += bs {
				spectrum := fft.Apply(
					int(h.SampleRate),
					data[i:i+bs],
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
		},
	)
}
