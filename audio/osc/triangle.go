package osc

// Triangle is an oscillator that writes a triangle wave of frequency `freq`, bit depth `depth`,
// and sample rate `sampleRate`, into the buffer of type T `buffer`
func Triangle[T BitDepths](buffer []T, freq, depth, sampleRate float64) {
	var wave = triangleWave(len(buffer), int(sampleRate), freq)

	for i := range buffer {
		buffer[i] = T(wave[i] * float64(int(2)<<int(depth-2)-1))
	}
}

func triangleWave(size, sampleRate int, freq float64) []float64 {
	var (
		sample    = -1.0
		buffer    = make([]float64, size)
		cycle     = float64(sampleRate) / freq
		halfCycle = cycle * 0.5
		step      = 1 / halfCycle
	)

	for i := 0; i < size; i++ {
		if i+1%int(halfCycle) == 0 {
			step = -step
		}

		sample += step
		buffer[i] = sample
	}

	return buffer

}
