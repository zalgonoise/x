//nolint:gomnd // hardcoded constant values, that are less readable if defined as constants
package fft

import (
	"math"

	"github.com/zalgonoise/x/audio/fft/window"
)

const (
	tau = 2 * math.Pi

	// DefaultMagnitudeThreshold describes the default value where a certain
	// frequency is strong enough to be considered relevant to the spectrum filter.
	DefaultMagnitudeThreshold = 10
)

// FrequencyPower denotes a single frequency and its magnitude in a Fast
// Fourier Transform of a signal.
type FrequencyPower struct {
	Freq int
	Mag  float64
}

// Apply applies a Fast Fourier Transform (FFT) on a slice of float64 `data`,
// with sample rate `sampleRate`. It returns a slice of FrequencyPower.
func Apply(sampleRate int, data []float64, w window.Window) []FrequencyPower {
	var (
		ln         = len(data)
		magnitudes = make([]FrequencyPower, 0, (ln/2)-1)
	)

	// apply a window function to the values
	if w != nil && len(w) == ln {
		w.Apply(data)
	}

	// apply a fast Fourier transform on the data; exclude index 0, no 0Hz-freq results
	spectrum := FFT(ToComplex(data))

	for i := 1; i < ln/2; i++ {
		freqReal := real(spectrum[i])
		freqImag := imag(spectrum[i])
		// map the magnitude for each frequency bin to the corresponding value in the map
		// using math.Sqrt(re*re + im*im) is faster than using math.Hypot(re, im)
		// see fft_test.go for details
		magnitudes = append(
			magnitudes,
			FrequencyPower{
				Freq: i * sampleRate / ln,
				Mag:  math.Sqrt(freqReal*freqReal + freqImag*freqImag),
			},
		)
	}

	return magnitudes
}

// FFT applies a Fast Fourier Transform to the input slice of complex128 values, to
// retrieve the frequency spectrum of a digital signal.
func FFT(value []complex128) []complex128 {
	var (
		valueLen = len(value)
		factors  = GetRadix2Factors(valueLen)
		temp     = make([]complex128, valueLen) // temp
	)

	value = ReorderData(value)

	// stage increases by a power of two
	for stage := 2; stage <= valueLen; stage <<= 1 {
		var (
			blocks      = valueLen / stage
			stage2Value = stage / 2
		)

		// iterate through each item in the batch, increasing by the stage value
		for batchIdx := 0; batchIdx < valueLen; batchIdx += stage {
			if stage == 2 { // "first stage" scenario
				var (
					reorderIdx  = value[batchIdx]
					reorderNext = value[batchIdx+1]
				)

				temp[batchIdx] = reorderIdx + reorderNext
				temp[batchIdx+1] = reorderIdx - reorderNext

				continue
			}

			for iter := 0; iter < stage2Value; iter++ {
				var (
					idx        = iter + batchIdx
					idx2       = idx + stage2Value
					reorderIdx = value[idx]
					factorized = value[idx2] * factors[blocks*iter]
				)

				temp[idx] = reorderIdx + factorized
				temp[idx2] = reorderIdx - factorized
			}
		}

		value, temp = temp, value
	}

	return value
}

// IFFT returns the Inverse Fast Fourier Transform of a given complex128 slice.
func IFFT(value []complex128) []complex128 {
	var (
		ln     = len(value)
		output = make([]complex128, ln)
		factor = complex(float64(ln), 0)
	)

	// Reverse inputs, which is calculated with modulo factor, hence value[0] as an outlier
	output[0] = value[0]

	for i, j := 1, ln-1; i < ln; i, j = i+1, j-1 {
		output[i] = value[j]
	}

	output = FFT(output)

	for i := range output {
		output[i] /= factor
	}

	return output
}

// Convolve returns the convolution of x ∗ y, applied to the complex128 slice x.
func Convolve(x, y []complex128) []complex128 {
	if len(x) != len(y) {
		return nil
	}

	x = FFT(x)
	y = FFT(y)

	for i := 0; i < len(x); i++ {
		x[i] *= y[i]
	}

	return IFFT(x)
}
