package conv

import (
	"unsafe"
)

const (
	maxInt24 float64 = 1<<23 - 1
	// minInt24 float64 = ^1<<22 + 1
)

// PCM24Bit is a 24bit audio Converter
type PCM24Bit struct{}

// Parse consumes the input audio buffer, returning its floating point audio representation
func (PCM24Bit) Parse(buf []byte) []float64 {
	buf32bit := copy24to32(buf)
	data := *(*[]int32)(unsafe.Pointer(&buf32bit))
	data = data[:len(buf32bit)/4]
	for i := range data {
		if data[i]&0x00800000 != 0 {
			data[i] |= ^0xffffff // handle signed integers
		}
	}

	return convert(
		data, func(v int32) float64 {
			return float64(v) / maxInt24
		},
	)
}

// Bytes consumes the input floating point audio buffer, returning its byte representation
func (PCM24Bit) Bytes(buf []float64) []byte {
	value := convert(
		buf, func(f float64) int32 {
			return int32(f * maxInt24)
		},
	)

	data := make([]byte, len(value)*3)
	for i := range value {
		append3Bytes(i, data, *(*[3]byte)(unsafe.Pointer(&value[i])))
	}

	return data
}

// Value consumes the input floating point audio buffer, returning its PCM audio values as a slice of int
func (PCM24Bit) Value(buf []float64) []int {
	return convert(
		buf, func(f float64) int {
			return int(f * maxInt24)
		},
	)
}
