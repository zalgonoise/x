package data

import (
	"unsafe"
)

const (
	maxInt32 float64 = 1<<31 - 1
	// minInt32 float64 = ^1<<30 + 1
)

var _ Converter = Conv32Bit{}

// Conv32Bit is a 32bit audio Converter
type Conv32Bit struct{}

// Parse consumes the input audio buffer, returning its floating point audio representation
func (Conv32Bit) Parse(buf []byte) []float64 {
	data := *(*[]int32)(unsafe.Pointer(&buf))

	return conv(
		data[:len(buf)/4], func(v int32) float64 {
			return float64(v) / maxInt32
		},
	)
}

// Bytes consumes the input floating point audio buffer, returning its byte representation
func (Conv32Bit) Bytes(buf []float64) []byte {
	value := conv(
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
func (Conv32Bit) Value(buf []float64) []int {
	return conv(
		buf, func(f float64) int {
			return int(f * maxInt32)
		},
	)
}
