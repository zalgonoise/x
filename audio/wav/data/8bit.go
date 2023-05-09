package data

import (
	"unsafe"
)

const (
	maxInt8 float64 = 1<<7 - 1
	// minInt8 float64 = ^1<<6 + 1
)

var _ = Converter(Conv8Bit{})

// Conv8Bit is a 8bit audio Converter
type Conv8Bit struct{}

// Parse consumes the input audio buffer, returning its floating point audio representation
func (c Conv8Bit) Parse(buf []byte) []float64 {
	return conv(
		*(*[]int8)(unsafe.Pointer(&buf)),
		func(v int8) float64 {
			return float64(v) / maxInt8
		},
	)
}

// Bytes consumes the input floating point audio buffer, returning its byte representation
func (c Conv8Bit) Bytes(buf []float64) []byte {
	value := conv(
		buf, func(f float64) int8 {
			return int8(f * maxInt8)
		},
	)

	return *(*[]byte)(unsafe.Pointer(&value))
}

// Value consumes the input floating point audio buffer, returning its PCM audio values as a slice of int
func (c Conv8Bit) Value(buf []float64) []int {
	return conv(
		buf, func(f float64) int {
			return int(f * maxInt8)
		},
	)
}
