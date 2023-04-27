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

// fullCycle provides a buffer length which allows a sine wave to complete a full cycle, until it reaches the same zero
// point where it started -- allowing for more precise sine waves, as they are generated for one cycle and copied over
//
// Annotations in the code show an example of a 2000Hz sine wave in a 44100Hz sample rate. A half period for this
// frequency is 22.05, and the function finds out that multiplying this value by 20 provides a rounded value of 441.
func fullCycle(sampleRate, freq float64) int {
	var (
		halfPeriod        = sampleRate / freq              // 44100 / 2000 == 22.05
		halfPeriodFloored = float64(int(halfPeriod))       // floored: 22
		halfPeriodDecimal = halfPeriod - halfPeriodFloored // decimals: 22.05 - 22 = 0.05000000000000071 (float64 drift)
	)

	if halfPeriodDecimal == 0 {
		return int(halfPeriod) // freq is multiple of sampleRate
	}

	// fix floating point drift
	halfPeriodDecimal = float64(int(halfPeriodDecimal*1000)) / 1000.0 // 0.05000000000000071 --> 0.05
	var mul = 1.0 / halfPeriodDecimal                                 // 1.0 / 0.05 == 20

	return int(halfPeriod * mul) // 22.05 * 20 == 441
}
