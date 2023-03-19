package osc

import (
	"math"
)

func Sine[T bitDepths](buffer []T, freq, depth, sampleRate float64) {
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

func sine[T bitDepths](buffer []T, freq, depth, sampleRate float64) {
	for i := 0; i < len(buffer); i++ {
		sample := math.Sin(2.0 * math.Pi * freq * float64(i) / sampleRate)
		buffer[i] = T(sample * (math.Pow(2.0, depth)/2.0 - 1.0))
	}
}
