package conv

import (
	"unsafe"
)

const (
	maxInt64    = 1<<63 - 1
	sizeFloat64 = 8
)

// Float64 is a 64bit IEEE Floating Point audio Converter.
//
// NOTE: This converter has not yet been tested and may present incorrect data, or raise errors.
type Float64 struct{}

// Parse consumes the input audio buffer, returning its floating point audio representation.
func (Float64) Parse(buf []byte) []float64 {
	data := *(*[]uint64)(unsafe.Pointer(&buf))

	return convert(
		data[:len(buf)/sizeFloat64], func(v uint64) float64 {
			return *(*float64)(unsafe.Pointer(&v))
		},
	)
}

// Bytes consumes the input floating point audio buffer, returning its byte representation.
func (Float64) Bytes(buf []float64) []byte {
	uintValue := *(*[]uint64)(unsafe.Pointer(&buf))

	data := make([]byte, len(uintValue)*sizeFloat64)

	for i := range buf {
		append8Bytes(i, data, *(*[sizeFloat64]byte)(unsafe.Pointer(&uintValue[i])))
	}

	return data
}

// Value consumes the input floating point audio buffer, returning its PCM audio values as a slice of int.
func (Float64) Value(buf []float64) []int {
	return convert(
		buf, func(f float64) int {
			return int(f * maxInt64)
		},
	)
}
