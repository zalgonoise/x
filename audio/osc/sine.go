package osc

import "github.com/zalgonoise/x/audio/trig"

// Sine is an oscillator that writes a sine wave of frequency `freq`, bit depth `depth`,
// and sample rate `sampleRate`, into the buffer of type T `buffer`
func Sine[T BitDepths](buffer []T, freq, depth, sampleRate int) {
	var wave = sine(len(buffer), sampleRate, freq)

	for i := range buffer {
		buffer[i] = T(wave[i] * float64(int(2)<<(depth-1)/2-1))
	}
}

func sine(size, sampleRate, freq int) []float64 {
	var (
		buffer      = make([]float64, size)
		tauOverFreq = tau * float64(freq)
	)

	for i := 0; i < size; i++ {
		buffer[i] = trig.Sin(tauOverFreq * (float64(i) / float64(sampleRate)))
	}

	return buffer
}
