package conv

import (
	"encoding/binary"
	"math"
)

func Float64To(v float64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, math.Float64bits(v))
	return buf
}

func Float64From(v []byte) float64 {
	if len(v) > 8 {
		return 0
	}
	return math.Float64frombits(binary.BigEndian.Uint64(v))
}

func Float32To(v float32) []byte {
	var buf = make([]byte, 4)
	binary.BigEndian.PutUint32(buf, math.Float32bits(v))
	return buf
}

func Float32From(v []byte) float32 {
	if len(v) > 4 {
		return 0
	}
	return math.Float32frombits(binary.BigEndian.Uint32(v))
}

func Uint64To(v uint64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, v)
	return buf
}

func Uint64From(v []byte) uint64 {
	if len(v) > 8 {
		return 0
	}
	return binary.BigEndian.Uint64(v)
}

func Uint32To(v uint32) []byte {
	var buf = make([]byte, 4)
	binary.BigEndian.PutUint32(buf, v)
	return buf
}

func Uint32From(v []byte) uint32 {
	if len(v) > 4 {
		return 0
	}
	return binary.BigEndian.Uint32(v)
}

func Uint16To(v uint16) []byte {
	var buf = make([]byte, 2)
	binary.BigEndian.PutUint16(buf, v)
	return buf
}

func Uint16From(v []byte) uint16 {
	if len(v) > 2 {
		return 0
	}
	return binary.BigEndian.Uint16(v)
}
