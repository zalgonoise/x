package conv

import (
	"unsafe"
)

const (
	maxInt32 float64 = 1<<31 - 1
	// minInt32 float64 = ^1<<30 + 1
)

// PCM32Bit is a 32bit audio Converter
type PCM32Bit struct{}

// Parse consumes the input audio buffer, returning its floating point audio representation
func (PCM32Bit) Parse(buf []byte) []float64 {
	data := *(*[]int32)(unsafe.Pointer(&buf))

	return convert(
		data[:len(buf)/4], func(v int32) float64 {
			return float64(v) / maxInt32
		},
	)
}

// Bytes consumes the input floating point audio buffer, returning its byte representation
func (PCM32Bit) Bytes(buf []float64) []byte {
	value := convert(
		buf, func(f float64) int32 {
			return int32(f * maxInt32)
		},
	)

	data := make([]byte, len(value)*4)
	for i := range value {
		append4Bytes(i, data, *(*[4]byte)(unsafe.Pointer(&value[i])))
	}

	return data
}

// Value consumes the input floating point audio buffer, returning its PCM audio values as a slice of int
func (PCM32Bit) Value(buf []float64) []int {
	return convert(
		buf, func(f float64) int {
			return int(f * maxInt32)
		},
	)
}
