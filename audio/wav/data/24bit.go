package data

import (
	"unsafe"
)

const (
	maxInt24 float64 = 1<<23 - 1
	// minInt24 float64 = ^1<<22 + 1
)

var _ = Converter(Conv24Bit{})

type Conv24Bit struct{}

func (c Conv24Bit) Parse(buf []byte) []float64 {
	buf32bit := copy24to32(buf)
	data := *(*[]int32)(unsafe.Pointer(&buf32bit))
	data = data[:len(buf32bit)/4]
	for i := range data {
		if data[i]&0x00800000 != 0 {
			data[i] |= ^0xffffff // handle signed integers
		}
	}

	return conv(
		data, func(v int32) float64 {
			return float64(v) / maxInt24
		},
	)
}

func (c Conv24Bit) Bytes(buf []float64) []byte {
	value := conv(
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

func (c Conv24Bit) Value(buf []float64) []int {
	return conv(
		buf, func(f float64) int {
			return int(f * maxInt24)
		},
	)
}
