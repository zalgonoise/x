package osc

import (
	"math"
)

// Sine is an oscillator that writes a sine wave of frequency `freq`, bit depth `depth`,
// and sample rate `sampleRate`, into the buffer of type T `buffer`
func Sine[T BitDepths](buffer []T, freq, depth, sampleRate float64) {
	halfPeriod := int(sampleRate / freq)

	if len(buffer) > halfPeriod {
		var wave = make([]T, halfPeriod)
		sine(wave, freq, depth, sampleRate)
		for i := 0; i < len(buffer); i += len(wave) {
			copy(buffer[i:], wave)
		}
		return
	}

	sine(buffer, freq, depth, sampleRate)
}

func sine[T BitDepths](buffer []T, freq, depth, sampleRate float64) {
	for i := 0; i < len(buffer); i++ {
		sample := math.Sin(tau * freq * float64(i) / sampleRate)
		buffer[i] = T(sample * float64(int(2)<<int(depth-1)/2-1))
	}
}
