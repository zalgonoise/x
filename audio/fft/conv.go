package fft

import (
	"unsafe"
)

// ToComplex converts a slice of float64 into a slice of complex128, where the input
// slice's values are set as the real part of the complex numbers
func ToComplex(x []float64) []complex128 {
	length := len(x)
	out := make([]float64, length<<1)
	for i := range x {
		out[i*2] = x[i]
	}
	return (*(*[]complex128)(unsafe.Pointer(&out)))[:length]
}
