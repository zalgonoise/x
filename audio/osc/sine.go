package osc

import "github.com/zalgonoise/x/audio/trig"

// Sine is an oscillator that writes a sine wave of frequency `freq`, bit depth `depth`,
// and sample rate `sampleRate`, into the buffer of type T `buffer`
func Sine[T BitDepths](buffer []T, freq, depth, sampleRate float64) {
	wave := sine(len(buffer), int(sampleRate), freq)

	for i := range buffer {
		buffer[i] = T(wave[i] * float64(int(2)<<int(depth-1)/2-1))
	}
}

func sine(size, sampleRate int, freq float64) []float64 {
	var buffer = make([]float64, size)

	for i := 0; i < size; i++ {
		buffer[i] = trig.Sin(tau * freq * (float64(i) / float64(sampleRate)))
	}

	return buffer
}
