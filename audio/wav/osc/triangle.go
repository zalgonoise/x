package osc

import (
	"math"
)

func Triangle[T bitDepths](buffer []T, freq, depth, sampleRate float64) {
	halfPeriod := int(sampleRate / freq)
	increment := 4.0 / float64(halfPeriod)
	var sampleInt T = -(1 << int(depth-1))

	if len(buffer) > halfPeriod {
		var wave = make([]T, halfPeriod)
		triangle(wave, halfPeriod, sampleInt, increment, depth)
		for i := 0; i < len(buffer); i += len(wave) {
			copy(buffer[i:], wave)
		}
		return
	}

	triangle(buffer, halfPeriod, sampleInt, increment, depth)
}

func triangle[T bitDepths](buffer []T, halfPeriod int, sampleInt T, increment, depth float64) {
	var swap bool
	for i := 0; i < len(buffer); i++ {
		if i%(halfPeriod/2) == 0 {
			swap = !swap
		}
		if swap {
			sampleInt += T(increment * (math.Pow(2.0, depth-1) - 1.0))
		} else {
			sampleInt -= T(increment * (math.Pow(2.0, depth-1) - 1.0))
		}
		buffer[i] = sampleInt
	}
}