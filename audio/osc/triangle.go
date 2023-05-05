package osc

// Triangle is an oscillator that writes a triangle wave of frequency `freq`, bit depth `depth`,
// and sample rate `sampleRate`, into the buffer of type T `buffer`
func Triangle[T BitDepths](buffer []T, freq, depth, sampleRate int) {
	var wave = buildFrom1Hz(len(buffer), sampleRate, freq, triangle1Hz)

	for i := range buffer {
		buffer[i] = T(wave[i] * float64(int(2)<<(depth-2)-1))
	}
}

func triangle1Hz(sampleRate int) []float64 {
	var (
		buffer    = make([]float64, sampleRate)
		halfCycle = sampleRate / 2
		stepValue = 2.0 / float64(sampleRate)
		sample    = -1.0
	)

	for i := 0; i < halfCycle; i++ {
		buffer[i] = sample
		sample += stepValue
	}

	for i, j := halfCycle, halfCycle-1; i < sampleRate; i, j = i+1, j-1 {
		buffer[i] = -buffer[j]
	}

	return buffer
}
