package conv

import (
	"unsafe"
)

// Float32 is a 32bit IEEE Floating Point audio Converter
type Float32 struct{}

// Parse consumes the input audio buffer, returning its floating point audio representation
func (Float32) Parse(buf []byte) []float64 {
	data := *(*[]uint32)(unsafe.Pointer(&buf))

	return convert(
		data[:len(buf)/4], func(v uint32) float64 {
			return float64(*(*float32)(unsafe.Pointer(&v)))
		},
	)
}

// Bytes consumes the input floating point audio buffer, returning its byte representation
func (Float32) Bytes(buf []float64) []byte {
	value := convert(
		buf, func(f float64) float32 {
			return float32(f)
		},
	)

	uintValue := *(*[]uint32)(unsafe.Pointer(&value))

	data := make([]byte, len(uintValue)*4)

	for i := range value {
		append4Bytes(i, data, *(*[4]byte)(unsafe.Pointer(&uintValue[i])))
	}

	return data

}

// Value consumes the input floating point audio buffer, returning its PCM audio values as a slice of int
func (Float32) Value(buf []float64) []int {
	return convert(
		buf, func(f float64) int {
			return int(f * maxInt32)
		},
	)
}