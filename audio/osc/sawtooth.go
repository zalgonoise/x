package osc

// Sawtooth is an oscillator that writes a sawtooth wave of frequency `freq`, bit depth `depth`,
// and sample rate `sampleRate`, into the buffer of type T `buffer`
func Sawtooth[T BitDepths](buffer []T, freq, depth, sampleRate float64) {
	var (
		halfPeriod = int(sampleRate / freq)
		increment  = 2.0 / float64(halfPeriod)
		sampleInt  T
		cycle      = fullCycle(sampleRate, freq)
	)

	if len(buffer) > halfPeriod {
		var wave = make([]T, halfPeriod)
		sawtooth(wave, halfPeriod, cycle, sampleInt, increment, depth)
		for i := 0; i < len(buffer); i += len(wave) {
			copy(buffer[i:], wave)
		}
		return
	}
	sawtooth(buffer, halfPeriod, cycle, sampleInt, increment, depth)
}

func sawtooth[T BitDepths](buffer []T, halfPeriod, cycle int, sampleInt T, increment, depth float64) {
	var base T = ^(2 << int(depth-2)) + 2
	inc := T(increment * float64(^base))

	for i := 0; i < len(buffer); i++ {
		if i+1%cycle == 0 {
			buffer[i] = sampleInt
			continue
		}

		if i%halfPeriod == 0 {
			sampleInt = base
		} else {
			sampleInt += inc
		}
		buffer[i] = sampleInt
	}
}
