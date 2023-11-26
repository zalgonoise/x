//nolint:dupl // similar to 32-bit's logic but different implementation of Converter
package conv

import (
	"unsafe"
)

const (
	maxInt16  float64 = 1<<15 - 1
	sizePCM16         = 2
)

// PCM16Bit is a 16bit audio Converter.
type PCM16Bit struct{}

// Parse consumes the input audio buffer, returning its floating point audio representation.
func (PCM16Bit) Parse(buf []byte) []float64 {
	data := *(*[]int16)(unsafe.Pointer(&buf))

	return convert(
		data[:len(buf)/sizePCM16],
		func(v int16) float64 {
			return float64(v) / maxInt16
		},
	)
}

// Bytes consumes the input floating point audio buffer, returning its byte representation.
func (PCM16Bit) Bytes(buf []float64) []byte {
	value := convert(
		buf, func(f float64) int16 {
			return int16(f * maxInt16)
		},
	)

	data := make([]byte, len(value)*sizePCM16)
	for i := range value {
		append2Bytes(i, data, *(*[sizePCM16]byte)(unsafe.Pointer(&value[i])))
	}

	return data
}

// Value consumes the input floating point audio buffer, returning its PCM audio values as a slice of int.
func (PCM16Bit) Value(buf []float64) []int {
	return convert(
		buf, func(f float64) int {
			return int(f * maxInt16)
		},
	)
}
