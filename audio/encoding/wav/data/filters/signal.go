package filters

// TODO: move to a different package outside of audio/encoding/wav/data

const sumElems = 2

// Add sums two signals together, repeating the pattern in `signal` if shorter than the
// audio buffer.
func Add(signal []float64) func([]float64) {
	return func(buffer []float64) {
		for i := range buffer {
			buffer[i] = (buffer[i] + signal[i%len(signal)]) / sumElems
		}
	}
}

// Sub subtracts two signals together, repeating the pattern in `signal` if shorter than the
// audio buffer.
func Sub(signal []float64) func([]float64) {
	return func(buffer []float64) {
		for i := range buffer {
			buffer[i] = (buffer[i] - signal[i%len(signal)]) / sumElems
		}
	}
}
