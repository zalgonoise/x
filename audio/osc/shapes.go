package osc

import (
	"math"
)

// Type is an enumeration for the supported oscillator types.
type Type uint8

const (
	// SineWave is the oscillator Type for a Sine Oscillator.
	SineWave Type = iota
	// SquareWave is the oscillator Type for a Square Oscillator.
	SquareWave
	// TriangleWave is the oscillator Type for a Triangle Oscillator.
	TriangleWave
	// SawtoothUpWave is the oscillator Type for a rising Sawtooth Oscillator.
	SawtoothUpWave
	// SawtoothDownWave is the oscillator Type for a falling Sawtooth Oscillator.
	SawtoothDownWave
)

func NewOscillator(waveType Type) Oscillator {
	switch waveType {
	case SineWave:
		return Sine
	case SquareWave:
		return Square
	case TriangleWave:
		return Triangle
	case SawtoothUpWave:
		return SawtoothUp
	case SawtoothDownWave:
		return SawtoothDown
	default:
		return nil
	}
}

const (
	tau float64 = math.Pi * 2
)

// Oscillator is a generic function that writes a wave of a certain shape into a buffer
// of float64 type.
type Oscillator func(buffer []float64, freq, depth, sampleRate int)

func buildFrom1Hz(size, sampleRate, freq int, oneHzFunc func(int) []float64) []float64 {
	var (
		buffer = make([]float64, size)
		base   = oneHzFunc(sampleRate)
	)

	for i, j := 0, 0; i < size; i, j = i+1, i*freq%sampleRate {
		buffer[i] = base[j]
	}

	return buffer
}
