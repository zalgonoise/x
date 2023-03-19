package osc

import (
	"math"
)

func Sawtooth[T bitDepths](buffer []T, freq, depth, sampleRate float64) {
	halfPeriod := int(sampleRate / freq)
	increment := 2.0 / float64(halfPeriod)
	var sampleInt T

	if len(buffer) > halfPeriod {
		var wave = make([]T, halfPeriod)
		sawtooth(wave, halfPeriod, sampleInt, increment, depth)
		for i := 0; i < len(buffer); i += len(wave) {
			copy(buffer[i:], wave)
		}
		return
	}
	sawtooth(buffer, halfPeriod, sampleInt, increment, depth)
}

func sawtooth[T bitDepths](buffer []T, halfPeriod int, sampleInt T, increment, depth float64) {
	for i := 0; i < len(buffer); i++ {
		if i%halfPeriod == 0 {
			sampleInt = -T(math.Pow(2.0, depth-1) - 1.0)
		}
		sampleInt += T(increment * (math.Pow(2.0, depth-1) - 1.0))
		buffer[i] = sampleInt
	}
}
