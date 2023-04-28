package osc

// Triangle is an oscillator that writes a triangle wave of frequency `freq`, bit depth `depth`,
// and sample rate `sampleRate`, into the buffer of type T `buffer`
func Triangle[T BitDepths](buffer []T, freq, depth, sampleRate float64) {
	var (
		halfPeriod, mul   = fullCycle(sampleRate, freq)
		gap               = halfPeriod * mul
		increment         = 4.0 / float64(halfPeriod)
		sampleInt       T = -(1 << int(depth-1))
	)

	if len(buffer) > halfPeriod {
		var wave = make([]T, halfPeriod)
		triangle(wave, halfPeriod, gap, sampleInt, increment, depth)

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

	triangle(buffer, halfPeriod, gap, sampleInt, increment, depth)
}

func triangle[T BitDepths](buffer []T, halfPeriod, gap int, sampleInt T, increment, depth float64) {
	var (
		swap          bool
		stepValue     = T(increment * float64(int(2)<<int(depth-2)-1))
		quarterPeriod = halfPeriod / 2
	)

	for i := 0; i < len(buffer); i++ {
		if i+1%gap == 0 {
			buffer[i] = sampleInt
			continue
		}

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
