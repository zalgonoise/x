package conv

import (
	"encoding/binary"
	"math"
)

func To64(v float64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, math.Float64bits(v))
	return buf
}

func From64(v []byte) float64 {
	if len(v) > 8 {
		return 0
	}

	return math.Float64frombits(binary.BigEndian.Uint64(v))
}

func To32(v float32) []byte {
	var buf = make([]byte, 4)
	binary.BigEndian.PutUint32(buf, math.Float32bits(v))
	return buf
}

func From32(v []byte) float32 {
	if len(v) > 4 {
		return 0
	}

	return math.Float32frombits(binary.BigEndian.Uint32(v))
}
