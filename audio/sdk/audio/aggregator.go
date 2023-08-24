package audio

import (
	"slices"

	"github.com/zalgonoise/x/audio/fft"
	"github.com/zalgonoise/x/audio/fft/window"
	"github.com/zalgonoise/x/audio/wav/header"
)

// Aggregator is a generic function type that serves as an audio processor function,
// but returns any type desired, as appropriate to the analysis, processing, recording, whatever it may be.
//
// It is of the responsibility of the Exporter to position the configured Aggregator to actually export the
// aggregations.
//
// The sole responsibility of an Aggregator is to convert raw audio (as chunks of float64 values) into anything
// meaningful, that is exported / handled separately. Not all Exporter will need one or more Aggregator, however
// these are supposed to be perceived as preset building blocks to work with the incoming audio chunks.
type Aggregator[T any] func(*header.Header, []float64) T

// MaxPeak returns a float64 Aggregator that calculates the maximum peak value in an audio signal
func MaxPeak() Aggregator[float64] {
	return func(_ *header.Header, data []float64) (maximum float64) {
		for i := range data {
			if data[i] > maximum {
				maximum = data[i]
			}
		}

		return maximum
	}
}

// AveragePeak returns a float64 Aggregator that calculates the average peak value in an audio signal
func AveragePeak() Aggregator[float64] {
	return func(_ *header.Header, data []float64) (average float64) {
		for i := range data {
			average += data[i]
		}

		return average / float64(len(data))
	}
}

// MaxSpectrum returns a []fft.FrequencyPower Aggregator that calculates the maximum spectrum values in an audio signal
func MaxSpectrum(size int) Aggregator[[]fft.FrequencyPower] {
	if size < 8 {
		size = 64
	}

	sampleRate := 44100

	return func(h *header.Header, data []float64) []fft.FrequencyPower {
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
	}
}
