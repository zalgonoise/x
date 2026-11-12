package fft

// ToComplex converts a slice of float64 into a slice of complex128, where the input
// slice's values are set as the real part of the complex numbers, and the imaginary parts are zero.
func ToComplex(x []float64) []complex128 {
	out := make([]complex128, len(x))

	for i := range x {
		out[i] = complex(x[i], 0)
	}

	return out
}
