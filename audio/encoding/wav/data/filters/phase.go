package filters

// TODO: move to a different package outside of audio/encoding/wav/data

// PhaseFlip inverts the phase of the input signal.
func PhaseFlip() func([]float64) {
	return func(buffer []float64) {
		for i := range buffer {
			buffer[i] = -buffer[i]
		}
	}
}
