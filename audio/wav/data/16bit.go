package data

import (
	"unsafe"
)

const (
	maxInt16 float64 = 1<<15 - 1
	// minInt16 float64 = ^1<<14 + 1
)

var _ = Converter(Conv16Bit{})

type Conv16Bit struct{}

func (c Conv16Bit) Parse(buf []byte) []float64 {
	data := *(*[]int16)(unsafe.Pointer(&buf))

	return conv(
		data[:len(buf)/2],
		func(v int16) float64 {
			return float64(v) / maxInt16
		},
	)
}

func (c Conv16Bit) Bytes(buf []float64) []byte {
	value := conv(
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

func (c Conv16Bit) Value(buf []float64) []int {
	return conv(
		buf, func(f float64) int {
			return int(f * maxInt16)
		},
	)
}
