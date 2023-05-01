package osc

// Square is an oscillator that writes a square wave of frequency `freq`, bit depth `depth`,
// and sample rate `sampleRate`, into the buffer of type T `buffer`
func Square[T BitDepths](buffer []T, freq, depth, sampleRate float64) {
	var (
		halfPeriod, mul = fullCycle(sampleRate, freq)
		gap             = halfPeriod * mul
		sampleInt       = T(2<<int16(depth-2)) - 1
	)

	if len(buffer) > halfPeriod {
		var wave = make([]T, halfPeriod)
		square(wave, halfPeriod, gap, sampleInt)

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

	square(buffer, halfPeriod, gap, sampleInt)
}

func square[T BitDepths](buffer []T, halfPeriod, gap int, sampleInt T) {
	quarterPeriod := halfPeriod / 2

	for i := 0; i < len(buffer); i++ {
		if i+1%gap == 0 {
			buffer[i] = buffer[i-1]
			i++
		}
		if i%halfPeriod < quarterPeriod {
			buffer[i] = sampleInt
			continue
		}

		buffer[i] = -sampleInt
	}
}
