package conv

import (
	"unsafe"
)

const (
	maxInt16 float64 = 1<<15 - 1
	// minInt16 float64 = ^1<<14 + 1
)

// Conv16Bit is a 16bit audio Converter
type Conv16Bit struct{}

// Parse consumes the input audio buffer, returning its floating point audio representation
func (Conv16Bit) Parse(buf []byte) []float64 {
	data := *(*[]int16)(unsafe.Pointer(&buf))

	return convert(
		data[:len(buf)/2],
		func(v int16) float64 {
			return float64(v) / maxInt16
		},
	)
}

// Bytes consumes the input floating point audio buffer, returning its byte representation
func (Conv16Bit) Bytes(buf []float64) []byte {
	value := convert(
		buf, func(f float64) int16 {
			return int16(f * maxInt16)
		},
	)

	data := make([]byte, len(value)*2)
	for i := range value {
		append2Bytes(i, data, *(*[2]byte)(unsafe.Pointer(&value[i])))
	}

	return data
}

// Value consumes the input floating point audio buffer, returning its PCM audio values as a slice of int
func (Conv16Bit) Value(buf []float64) []int {
	return convert(
		buf, func(f float64) int {
			return int(f * maxInt16)
		},
	)
}
