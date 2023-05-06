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
		buffer       = make([]float64, sampleRate)
		halfCycle    = sampleRate / 2
		quarterCycle = halfCycle / 2
		stepValue    = 1.0 / float64(quarterCycle) // from -1.0 to 0.0
	)

	// gen first quadrant (Q1)
	for i, j := 0, -1.0; i < quarterCycle; i, j = i+1, j+stepValue {
		buffer[i] = j
	}
	buffer[quarterCycle] = 0

	// Q2 is negative reverse of Q2
	for i, j := quarterCycle+1, quarterCycle-1; i < halfCycle; i, j = i+1, j-1 {
		buffer[i] = -buffer[j]
	}
	buffer[halfCycle] = 1.0

	// Q3+Q4 are reverse of Q2+Q1
	for i, j := halfCycle+1, halfCycle-1; i < sampleRate; i, j = i+1, j-1 {
		buffer[i] = buffer[j]
	}

	return buffer
}
