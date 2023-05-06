package osc

// SawtoothUp is an oscillator that writes a rising sawtooth wave of frequency `freq`, bit depth `depth`,
// and sample rate `sampleRate`, into the buffer of type T `buffer`
func SawtoothUp[T BitDepths](buffer []T, freq, depth, sampleRate int) {
	var wave = buildFrom1Hz(len(buffer), sampleRate, freq, sawtoothUp1Hz)

	for i := range buffer {
		buffer[i] = T(wave[i] * float64(int(2)<<(depth-2)-1))
	}
}

// SawtoothDown is an oscillator that writes a falling sawtooth wave of frequency `freq`, bit depth `depth`,
// and sample rate `sampleRate`, into the buffer of type T `buffer`
func SawtoothDown[T BitDepths](buffer []T, freq, depth, sampleRate int) {
	var wave = buildFrom1Hz(len(buffer), sampleRate, freq, sawtoothDown1Hz)

	for i := range buffer {
		buffer[i] = T(wave[i] * float64(int(2)<<(depth-2)-1))
	}
}

func sawtoothUp1Hz(sampleRate int) []float64 {
	var (
		buffer    = make([]float64, sampleRate)
		halfCycle = sampleRate / 2
		stepValue = 1.0 / float64(halfCycle) // from -1.0 to 0.0
	)

	for i, j := 0, -1.0; i < halfCycle; i, j = i+1, j+stepValue {
		buffer[i] = j
	}
	buffer[halfCycle] = 0

	// Q3+Q4 are negative reverse of Q2+Q1
	for i, j := halfCycle+1, halfCycle-1; i < sampleRate; i, j = i+1, j-1 {
		buffer[i] = -buffer[j]
	}

	return buffer
}

func sawtoothDown1Hz(sampleRate int) []float64 {
	var (
		buffer    = make([]float64, sampleRate)
		halfCycle = sampleRate / 2
		stepValue = 1.0 / float64(halfCycle) // from -1.0 to 0.0
	)

	// gen first two quadrants (Q1+Q2)
	for i, j := 0, 1.0; i < halfCycle; i, j = i+1, j-stepValue {
		buffer[i] = j
	}
	buffer[halfCycle] = 0

	// Q3+Q4 are negative reverse of Q2+Q1
	for i, j := halfCycle+1, halfCycle-1; i < sampleRate; i, j = i+1, j-1 {
		buffer[i] = -buffer[j]
	}

	return buffer
}
