package osc

// Sawtooth is an oscillator that writes a sawtooth wave of frequency `freq`, bit depth `depth`,
// and sample rate `sampleRate`, into the buffer of type T `buffer`
func Sawtooth[T BitDepths](buffer []T, freq, depth, sampleRate int) {
	var wave = buildFrom1Hz(len(buffer), sampleRate, freq, sawtooth1Hz)

	for i := range buffer {
		buffer[i] = T(wave[i] * float64(int(2)<<(depth-2)-1))
	}
}

func sawtooth1Hz(sampleRate int) []float64 {
	var (
		buffer    = make([]float64, sampleRate)
		stepValue = 2.0 / float64(sampleRate) // from -1.0 to +1.0
		sample    = -1.0
	)

	for i := 0; i < sampleRate; i++ {
		buffer[i] = sample
		sample += stepValue
	}

	return buffer
}
