//nolint:gomnd // contains hardcoded constants, would be less readable to declare these values as such
package osc

// Square is an oscillator that writes a square wave of frequency `freq`, bit-depth `depth`,
// and sample rate `sampleRate`, into the buffer of type T `buffer`.
func Square(buffer []float64, freq, depth, sampleRate int) {
	wave := buildFrom1Hz(len(buffer), sampleRate, freq, square1Hz)

	for i := range buffer {
		buffer[i] = wave[i] * float64(int(2)<<(depth-2)-1)
	}
}

func square1Hz(sampleRate int) []float64 {
	var (
		buffer    = make([]float64, sampleRate)
		halfCycle = sampleRate / 2
	)

	for i := 0; i < halfCycle; i++ {
		buffer[i] = 1.0
	}

	for i := halfCycle; i < sampleRate; i++ {
		buffer[i] = -1.0
	}

	return buffer
}
