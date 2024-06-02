package lpc

import (
	"math"
)

const (
	exp64 = 63
	exp32 = 31
	exp16 = 15
	exp8  = 7
)

func GolombEncode64(x, m uint64) (q, r uint64, ok bool) {
	if m == 0 {
		return uint64(math.Log2(float64(x) + 1.0)), x + 1, true
	}

	if m > exp64 {
		return 0, 0, false
	}

	var shift uint64 = 1 << m

	return uint64(math.Log2(float64(x/shift) + 1.0)), x - shift, true
}

func GolombDecode64(m, r uint64) (x uint64, ok bool) {
	switch {
	case m < 0:
		return 0, false
	case m > exp64:
		return 0, false
	case m == 0:
		return r - 1, true
	default:
		return r + (1 << m), true
	}
}

func GolombEncode32(x, m uint32) (q, r uint32, ok bool) {
	if m == 0 {
		return uint32(math.Log2(float64(x) + 1.0)), x + 1, true
	}

	if m > exp32 {
		return 0, 0, false
	}

	var shift uint32 = 1 << m

	return uint32(math.Log2(float64(x/shift) + 1.0)), x - shift, true
}

func GolombDecode32(m, r uint32) (x uint32, ok bool) {
	switch {
	case m < 0:
		return 0, false
	case m > exp32:
		return 0, false
	case m == 0:
		return r - 1, true
	default:
		return r + (1 << m), true
	}
}

func GolombEncode16(x, m uint16) (q, r uint16, ok bool) {
	if m == 0 {
		return uint16(math.Log2(float64(x) + 1.0)), x + 1, true
	}

	if m > exp16 {
		return 0, 0, false
	}

	var shift uint16 = 1 << m

	return uint16(math.Log2(float64(x/shift) + 1.0)), x - shift, true
}

func GolombDecode16(m, r uint16) (x uint16, ok bool) {
	switch {
	case m < 0:
		return 0, false
	case m > exp16:
		return 0, false
	case m == 0:
		return r - 1, true
	default:
		return r + (1 << m), true
	}
}

func GolombEncode8(x, m uint8) (q, r uint8, ok bool) {
	if m == 0 {
		return uint8(math.Log2(float64(x) + 1.0)), x + 1, true
	}

	if m > exp8 {
		return 0, 0, false
	}

	var shift uint8 = 1 << m

	return uint8(math.Log2(float64(x/shift) + 1.0)), x - shift, true
}

func GolombDecode8(m, r uint8) (x uint8, ok bool) {
	switch {
	case m < 0:
		return 0, false
	case m > exp8:
		return 0, false
	case m == 0:
		return r - 1, true
	default:
		return r + (1 << m), true
	}
}

type GolombWriter struct {
	w BitWriter
	m int
}

func (w *GolombWriter) WriteInt8(v int8) {
	var value uint8

	switch {
	case v == 0:
	case v < 0:
		value = 2 * uint8(-v)
	default:
		value = (2 * uint8(v)) - 1
	}

	// TODO: use r value here:
	q, _, ok := GolombEncode8(value, uint8(w.m))
	if !ok {
		return
	}

	mantissa := make([]bool, q+1)
	mantissa[len(mantissa)-1] = true

	// TODO: short binary of r
	// TODO: write mantissa + r to buffer
}

func bitLength(value uint8) int {
	bits := 8

	for i := uint8(1 << 7); i > 0; i = i >> 1 {
		if i&value == 0 {
			bits--

			continue
		}

		break
	}

	return bits
}
