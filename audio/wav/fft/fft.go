package fft

import (
	"math"

	"github.com/mjibson/go-dsp/fft"
)

// FrequencyPower denotes a single frequency and its magnitude in a Fast
// Fourier Transform of a signal
type FrequencyPower struct {
	Freq int
	Mag  float64
}

// BlockSize is an enumeration for FFT BlockSize values, which are a power of 2,
// from 8 to 8192
type BlockSize int

const (
	_ BlockSize = 1 << iota
	_
	_
	Block8
	Block16
	Block32
	Block64
	Block128
	Block256
	Block512
	Block1024
	Block2048
	Block4096
	Block8192
)

const (
	tau = math.Pi * 2

	// DefaultMagnitudeThreshold describes the default value where a certain
	// frequency is strong enough to be considered relevant to the spectrum filter
	DefaultMagnitudeThreshold = 10
)

func hamming(n int) []float64 {
	w := make([]float64, n)
	for i := 0; i < n; i++ {
		w[i] = 0.54 - 0.46*math.Cos(tau*float64(i)/(float64(n)-1))
	}
	return w
}

// Compute applies a Fast Fourier Transform (FFT) on a slice of float64 `data`,
// with sample rate `sampleRate`. It returns a slice of FrequencyPower
func Compute(sampleRate int, data []float64) []FrequencyPower {
	var (
		n          = len(data)
		freqUnit   = sampleRate / n
		magnitudes = make([]FrequencyPower, 0, (n/2)-1)
	)

	// apply a window function to the values
	window := hamming(n)
	for i := 0; i < n; i++ {
		data[i] *= window[i]
	}

	// apply a fast Fourier transform on the data; exclude index 0, no 0Hz-freq results
	frequencies := fft.FFTReal(data)
	for i := 1; i < n/2; i++ {
		freqReal := real(frequencies[i])
		freqImag := imag(frequencies[i])
		// map the magnitude for each frequency bin to the corresponding value in the map
		magnitudes = append(
			magnitudes,
			FrequencyPower{
				Freq: i * freqUnit,
				Mag:  math.Sqrt(freqReal*freqReal + freqImag*freqImag),
			},
		)
	}
	return magnitudes
}