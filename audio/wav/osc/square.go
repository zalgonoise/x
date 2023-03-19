package osc

import (
	"math"
)

func Square[T bitDepths](buffer []T, freq, depth, sampleRate float64) {
	halfPeriod := int(sampleRate / (2.0 * freq))
	sampleInt := T(math.Pow(2.0, depth-1) - 1.0)

	if len(buffer) > halfPeriod {
		var wave = make([]T, halfPeriod)
		square(wave, halfPeriod, sampleInt)
		for i := 0; i < len(buffer); i += len(wave) {
			copy(buffer[i:], wave)
		}
		return
	}

	square(buffer, halfPeriod, sampleInt)
}

func square[T bitDepths](buffer []T, halfPeriod int, sampleInt T) {
	for i := 0; i < len(buffer); i++ {
		if i%halfPeriod < halfPeriod/2 {
			buffer[i] = sampleInt
			continue
		}
		buffer[i] = -sampleInt
	}
}
