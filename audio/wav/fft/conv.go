package fft

import (
	"unsafe"
)

func ToComplex(x []float64) []complex128 {
	length := len(x)
	out := make([]float64, length<<1)
	for i := range x {
		out[i*2] = x[i]
	}
	return (*(*[]complex128)(unsafe.Pointer(&out)))[:length]
}
