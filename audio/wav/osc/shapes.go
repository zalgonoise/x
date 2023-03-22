package osc

import (
	"math"
)

// Type is an enumeration for the supported oscillator types
type Type uint8

const (
	// SineWave is the oscillator Type for a Sine Oscillator
	SineWave Type = iota
	// SquareWave is the oscillator Type for a Square Oscillator
	SquareWave
	// TriangleWave is the oscillator Type for a Triangle Oscillator
	TriangleWave
	// SawtoothWave is the oscillator Type for a Sawtooth Oscillator
	SawtoothWave
)

const (
	tau float64 = math.Pi * 2
)

// BitDepths describes the type constraint for Oscillator functions, for specific
// bit depths
type BitDepths interface {
	int8 | int16 | int32
}

// Oscillator is a generic function that writes a wave of a certain shape into a buffer
// of BitDepths type
type Oscillator[T BitDepths] func(buffer []T, freq, depth, sampleRate float64)
