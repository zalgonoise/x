package data

import (
	"unsafe"
)

const (
	maxInt8 float64 = 1<<7 - 1
	// minInt8 float64 = ^1<<6 + 1
)

var _ = Converter(Conv8Bit{})

type Conv8Bit struct{}

func (c Conv8Bit) Parse(buf []byte) []float64 {
	return conv(
		*(*[]int8)(unsafe.Pointer(&buf)),
		func(v int8) float64 {
			return float64(v) / maxInt8
		},
	)
}

func (c Conv8Bit) Bytes(buf []float64) []byte {
	value := conv(
		buf, func(f float64) int8 {
			return int8(f * maxInt8)
		},
	)

	return *(*[]byte)(unsafe.Pointer(&value))
}

func (c Conv8Bit) Value(buf []float64) []int {
	return conv(
		buf, func(f float64) int {
			return int(f * maxInt8)
		},
	)
}
