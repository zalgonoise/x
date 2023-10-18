package filters

// Amplify will transform the signal's intensity according tto the input ratio.
//
// Clipping (cutting off the floating point audio data at +1.0 and -1.0) may occur
// if the signal is amplified beyond the acceptable range, of -1.0 and +1.0.
//
// For an approach to Amplify that avoids clipping, use Normalize
func Amplify(ratio float64) func([]float64) {
	return func(buffer []float64) {
		for i := range buffer {
			buffer[i] = buffer[i] * ratio

			if buffer[i] > 1.0 {
				buffer[i] = 1.0
			} else if buffer[i] < -1.0 {
				buffer[i] = -1.0
			}
		}
	}
}

// Normalize amplifies the signal as much as possible before hitting
// peak levels in the signal.
//
// This is done by measuring the peak value of the signal and amplifying it
// by the ratio of (1.0 + 1 / maxValue)
func Normalize() func([]float64) {
	return func(buffer []float64) {
		var maxValue float64

		// find peak value, as a positive float64 value
		for _, v := range buffer {
			if v < 0.0 {
				v = -v
			}

			if v > maxValue {
				maxValue = v
			}
		}

		// normalize the signal, amplifying it as much as possible
		// before reaching peak levels
		Amplify(1.0 + 1/maxValue)(buffer)
	}
}
