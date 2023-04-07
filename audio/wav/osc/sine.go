package osc

import (
	"math"
)

// Sine is an oscillator that writes a sine wave of frequency `freq`, bit depth `depth`,
// and sample rate `sampleRate`, into the buffer of type T `buffer`
func Sine[T BitDepths](buffer []T, freq, depth, sampleRate float64) {
	cycle := fullCycle(sampleRate, freq)

	if len(buffer) > cycle {
		var wave = make([]T, cycle)
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
