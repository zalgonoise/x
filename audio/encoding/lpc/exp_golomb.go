package lpc

import "math"

func GolombEncode64(x, m uint64) (q, r uint64, ok bool) {
	if m == 0 || m > 63 {
		return 0, 0, false
	}

	var shift uint64 = 1 << m

	return uint64(math.Log2(float64(x/shift) + 1.0)), x - shift, true
}

func GolombDecode64(m, r uint64) (x uint64, ok bool) {
	if m < 1 {
		return 0, false
	}

	return r + (1 << m), true
}

func GolombEncode32(x, m uint32) (q, r uint32, ok bool) {
	if m == 0 || m > 31 {
		return 0, 0, false
	}

	var shift uint32 = 1 << m

	return uint32(math.Log2(float64(x/shift) + 1.0)), x - shift, true
}

func GolombDecode32(m, r uint32) (x uint32, ok bool) {
	if m < 1 {
		return 0, false
	}

	return r + (1 << m), true
}

func GolombEncode16(x, m uint16) (q, r uint16, ok bool) {
	if m == 0 || m > 15 {
		return 0, 0, false
	}

	var shift uint16 = 1 << m

	return uint16(math.Log2(float64(x/shift) + 1.0)), x - shift, true
}

func GolombDecode16(m, r uint16) (x uint16, ok bool) {
	if m < 1 {
		return 0, false
	}

	return r + (1 << m), true
}

func GolombEncode8(x, m uint8) (q, r uint8, ok bool) {
	if m == 0 || m > 7 {
		return 0, 0, false
	}

	var shift uint8 = 1 << m

	return uint8(math.Log2(float64(x/shift) + 1.0)), x - shift, true
}

func GolombDecode8(m, r uint8) (x uint8, ok bool) {
	if m < 1 {
		return 0, false
	}

	return r + (1 << m), true
}
