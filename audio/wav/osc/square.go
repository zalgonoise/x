package osc

// Square is an oscillator that writes a square wave of frequency `freq`, bit depth `depth`,
// and sample rate `sampleRate`, into the buffer of type T `buffer`
func Square[T BitDepths](buffer []T, freq, depth, sampleRate float64) {
	halfPeriod := int(sampleRate) / (2 * int(freq))
	sampleInt := T(2<<int16(depth-2)) - 1

	if len(buffer) > halfPeriod {
		var wave = make([]T, halfPeriod)
		square(wave, halfPeriod, sampleInt)
		for i := 0; i < len(buffer); i += len(wave) {
			copy(buffer[i:], wave)
		}
		return
	}

	square(buffer, halfPeriod, sampleInt)
}

func square[T BitDepths](buffer []T, halfPeriod int, sampleInt T) {
	quarterPeriod := halfPeriod / 2

	for i := 0; i < len(buffer); i++ {
		if i%halfPeriod < quarterPeriod {
			buffer[i] = sampleInt
			continue
		}
		buffer[i] = -sampleInt
	}
}
