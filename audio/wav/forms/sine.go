package forms

import (
	"math"
)

type Type uint8

const (
	SineWave Type = iota
	SquareWave
	TriangleWave
	SawtoothWave
)

type bitDepths interface {
	int8 | int16 | int32
}

type FormFunc[T bitDepths] func(buffer []T, freq, depth, sampleRate float64)

func Sine[T bitDepths](buffer []T, freq, depth, sampleRate float64) {
	for i := 0; i < len(buffer); i++ {
		sample := math.Sin(2.0 * math.Pi * freq * float64(i) / sampleRate)
		buffer[i] = T(sample * (math.Pow(2.0, depth)/2.0 - 1.0))
	}
}
