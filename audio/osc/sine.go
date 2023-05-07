package osc

import "github.com/zalgonoise/x/audio/trig"

// Sine is an oscillator that writes a sine wave of frequency `freq`, bit depth `depth`,
// and sample rate `sampleRate`, into the buffer of type T `buffer`
func Sine(buffer []float64, freq, depth, sampleRate int) {
	var wave = buildFrom1Hz(len(buffer), sampleRate, freq, sine1Hz)

	for i := range buffer {
		buffer[i] = wave[i] * float64(int(2)<<(depth-1)/2-1)
	}
}

func sine1Hz(sampleRate int) []float64 {
	var (
		buffer       = make([]float64, sampleRate)
		halfCycle    = sampleRate / 2
		quarterCycle = halfCycle / 2
	)

	// gen first quadrant (Q1)
	for i := 0; i < quarterCycle; i++ {
		buffer[i] = trig.Sin(tau * float64(i) / float64(sampleRate))
	}
	buffer[quarterCycle] = 1.0

	// Q2 is reverse of Q2
	for i, j := quarterCycle+1, quarterCycle-1; i < halfCycle; i, j = i+1, j-1 {
		buffer[i] = buffer[j]
	}
	buffer[halfCycle] = 0.0

	// Q3+Q4 are negative reverse of Q2+Q1
	for i, j := halfCycle+1, halfCycle-1; i < sampleRate; i, j = i+1, j-1 {
		buffer[i] = -buffer[j]
	}

	return buffer
}
