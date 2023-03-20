package osc

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

const (
	tau float64 = math.Pi * 2
)

type bitDepths interface {
	int8 | int16 | int32
}

type Oscillator[T bitDepths] func(buffer []T, freq, depth, sampleRate float64)
