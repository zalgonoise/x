package lpc

import "math"

// EliasGammaUint8 encodes the uint8 value in Elias Gamma code
//
// ref: https://en.wikipedia.org/wiki/Elias_gamma_coding
func EliasGammaUint8(value uint8) []bool {
	if value == 0 {
		return nil
	}

	log := int(math.Log2(float64(value)))

	code := make([]bool, log+1, log+8)
	code[len(code)-1] = true

	bits := asBits(value % (uint8(1) << log))

	return append(code, bits[len(bits)-log:]...)
}
