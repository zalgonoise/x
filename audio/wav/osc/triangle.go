package osc

// Triangle is an oscillator that writes a triangle wave of frequency `freq`, bit depth `depth`,
// and sample rate `sampleRate`, into the buffer of type T `buffer`
func Triangle[T BitDepths](buffer []T, freq, depth, sampleRate float64) {
	halfPeriod := int(sampleRate / freq)
	increment := 4.0 / float64(halfPeriod)
	var sampleInt T = -(1 << int(depth-1))

	if len(buffer) > halfPeriod {
		var wave = make([]T, halfPeriod)
		triangle(wave, halfPeriod, sampleInt, increment, depth)
		for i := 0; i < len(buffer); i += len(wave) {
			copy(buffer[i:], wave)
		}
		return
	}

	triangle(buffer, halfPeriod, sampleInt, increment, depth)
}

func triangle[T BitDepths](buffer []T, halfPeriod int, sampleInt T, increment, depth float64) {
	var (
		swap          bool
		stepValue     = T(increment * float64(int(2)<<int(depth-2)-1))
		quarterPeriod = halfPeriod / 2
	)
	for i := 0; i < len(buffer); i++ {
		if i%(quarterPeriod) == 0 {
			swap = !swap
		}
		if swap {
			sampleInt += stepValue
		} else {
			sampleInt -= stepValue
		}
		buffer[i] = sampleInt
	}
}
