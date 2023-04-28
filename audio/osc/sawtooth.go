package osc

// Sawtooth is an oscillator that writes a sawtooth wave of frequency `freq`, bit depth `depth`,
// and sample rate `sampleRate`, into the buffer of type T `buffer`
func Sawtooth[T BitDepths](buffer []T, freq, depth, sampleRate float64) {
	var (
		halfPeriod, mul = fullCycle(sampleRate, freq)
		gap             = halfPeriod * mul
		increment       = 2.0 / float64(halfPeriod)
		sampleInt       T
	)

	if len(buffer) > halfPeriod {
		var wave = make([]T, halfPeriod)
		sawtooth(wave, halfPeriod, gap, sampleInt, increment, depth)

		for i, j := 0, 0; i < len(buffer); i, j = i+len(wave), j+1 {
			copy(buffer[i:], wave)

			next := i + len(wave)
			if j+1%mul == 0 && next < len(buffer) {
				buffer[next+1] = buffer[next]
				i++
			}
		}
		return
	}

	sawtooth(buffer, halfPeriod, gap, sampleInt, increment, depth)
}

func sawtooth[T BitDepths](buffer []T, halfPeriod, gap int, sampleInt T, increment, depth float64) {
	var base T = ^(2 << int(depth-2)) + 2
	inc := T(increment * float64(^base))

	for i := 0; i < len(buffer); i++ {
		if i+1%gap == 0 {
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
