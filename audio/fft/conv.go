//nolint:gomnd // contains hardcoded values; makes code less readable to make these constants
package fft

import (
	"unsafe"
)

// ToComplex converts a slice of float64 into a slice of complex128, where the input
// slice's values are set as the real part of the complex numbers
//
// In Go, a complex number is a pair of two float numbers. This means that converting a
// float number to a complex number only makes it necessary to assign the (float's) value as
// the real part of the complex number, leaving the imaginary as zero.
//
// This particular function leverages that by creating a float slice of twice the size as the inputs, and
// placing each element of the input on every second slot. Finally, it casts this float slice as a complex slice,
// using an unsafe approach to prioritize performance.
func ToComplex(x []float64) []complex128 {
	var (
		length = len(x)
		out    = make([]float64, length*2)
	)

	for i, j := 0, 0; i < len(x); i, j = i+1, (i+1)*2 {
		out[j] = x[i]
	}

	return (*(*[]complex128)(unsafe.Pointer(&out)))[:length]
}
