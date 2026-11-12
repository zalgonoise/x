package filters

// TODO: move to a different package outside of audio/encoding/wav/data

const sumElems = 2

// Add mixes two signals together as their average ((a + b) / 2, i.e. a sum with 6 dB of headroom so the result
// cannot clip), repeating the pattern in `signal` if shorter than the audio buffer. An empty signal is a no-op.
func Add(signal []float64) func([]float64) {
	return func(buffer []float64) {
		if len(signal) == 0 {
			return
		}

		for i := range buffer {
			buffer[i] = (buffer[i] + signal[i%len(signal)]) / sumElems
		}
	}
}

// Sub subtracts `signal` from the buffer as half their difference ((a - b) / 2, so the result cannot clip), repeating
// the pattern in `signal` if shorter than the audio buffer. An empty signal is a no-op.
func Sub(signal []float64) func([]float64) {
	return func(buffer []float64) {
		if len(signal) == 0 {
			return
		}

		for i := range buffer {
			buffer[i] = (buffer[i] - signal[i%len(signal)]) / sumElems
		}
	}
}
