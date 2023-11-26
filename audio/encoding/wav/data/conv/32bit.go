//nolint:dupl // similar to 16-bit's logic but different implementation of Converter
package conv

import (
	"unsafe"
)

const (
	maxInt32  float64 = 1<<31 - 1
	sizePCM32         = 4
)

// PCM32Bit is a 32bit audio Converter.
type PCM32Bit struct{}

// Parse consumes the input audio buffer, returning its floating point audio representation.
func (PCM32Bit) Parse(buf []byte) []float64 {
	data := *(*[]int32)(unsafe.Pointer(&buf))

	return convert(
		data[:len(buf)/sizePCM32], func(v int32) float64 {
			return float64(v) / maxInt32
		},
	)
}

// Bytes consumes the input floating point audio buffer, returning its byte representation.
func (PCM32Bit) Bytes(buf []float64) []byte {
	value := convert(
		buf, func(f float64) int32 {
			return int32(f * maxInt32)
		},
	)

	data := make([]byte, len(value)*sizePCM32)
	for i := range value {
		append4Bytes(i, data, *(*[sizePCM32]byte)(unsafe.Pointer(&value[i])))
	}

	return data
}

// Value consumes the input floating point audio buffer, returning its PCM audio values as a slice of int.
func (PCM32Bit) Value(buf []float64) []int {
	return convert(
		buf, func(f float64) int {
			return int(f * maxInt32)
		},
	)
}
