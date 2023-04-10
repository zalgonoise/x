package fft

import (
	"math"

	dspfft "github.com/mjibson/go-dsp/fft"
	"github.com/zalgonoise/x/audio/wav/fft/window"
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

var blockSizeMap = map[int]BlockSize{
	8:    Block8,
	16:   Block16,
	32:   Block32,
	64:   Block64,
	128:  Block128,
	256:  Block256,
	512:  Block512,
	1024: Block1024,
	2048: Block2048,
	4096: Block4096,
	8192: Block8192,
}

// AsBlock returns a valid BlockSize for the input int `size`. If the input
// size is not valid, a default BlockSize is returned (Block1024)
func AsBlock(size int) BlockSize {
	if bs, ok := blockSizeMap[size]; ok {
		return bs
	}
	return Block1024
}

// DefaultMagnitudeThreshold describes the default value where a certain
// frequency is strong enough to be considered relevant to the spectrum filter
const DefaultMagnitudeThreshold = 10

// Apply applies a Fast Fourier Transform (FFT) on a slice of float64 `data`,
// with sample rate `sampleRate`. It returns a slice of FrequencyPower
func Apply(sampleRate int, data []float64, w window.Window) []FrequencyPower {
	var (
		n          = len(data)
		freqUnit   = sampleRate / n
		magnitudes = make([]FrequencyPower, 0, (n/2)-1)
	)

	// apply a window function to the values
	if w != nil && len(w) == n {
		w.Apply(data)
	}

	// apply a fast Fourier transform on the data; exclude index 0, no 0Hz-freq results
	frequencies := dspfft.FFTReal(data)
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
