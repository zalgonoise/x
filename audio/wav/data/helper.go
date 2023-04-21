package data

import "fmt"

// BitDepthTypes is a type constraint joining all the different
// data types used by the supported bit depths
type BitDepthTypes interface {
	int8 | int16 | int32 | byte | int | float64
}

func conv[F, T BitDepthTypes](from []F, fn func(F) T) []T {
	out := make([]T, len(from))
	for i := range from {
		out[i] = fn(from[i])
	}
	return out
}

func to[F, T BitDepthTypes](from []F) []T {
	out := make([]T, len(from))
	for i := range from {
		out[i] = (T)(from[i])
	}
	return out
}

func copy24to32(b []byte) []byte {
	out := make([]byte, len(b)+len(b)/3)

	for i, j := 0, 1; i < len(b); i, j = i+3, j+4 {

		fmt.Printf("added -- %v\n", b[i:i+3])
		copy(out[j:], b[i:i+3])
	}
	return out
}

// can't inline a pointer cast and convert an array to a slice
func append2Bytes(idx int, dst []byte, src [2]byte) {
	if idx*2 < len(dst) {
		copy(dst[idx*2:], src[:])
	}
}

// can't inline a pointer cast and convert an array to a slice
func append3Bytes(idx int, dst []byte, src [3]byte) {
	if idx*3 < len(dst) {
		copy(dst[idx*3:], src[:])
	}
}

// can't inline a pointer cast and convert an array to a slice
func append4Bytes(idx int, dst []byte, src [4]byte) {
	if idx*4 < len(dst) {
		copy(dst[idx*4:], src[:])
	}
}
