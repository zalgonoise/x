package fft

import (
	"github.com/zalgonoise/x/audio/trig"
)

const (
	firstCur  = 2048
	firstPrev = 1024
)

// GetRadix2Factors is temporarily public, could become private at a later point.,
func GetRadix2Factors(inputLen int) []complex128 {
	if factors, ok := radix2Factors[inputLen]; ok {
		return factors
	}

	for factor, prev := firstCur, firstPrev; factor <= inputLen; factor, prev = factor<<1, factor {
		if _, ok := radix2Factors[factor]; !ok {
			radix2Factors[factor] = make([]complex128, factor)
			for n, j := 0, 0; n < factor; n, j = n+2, j+1 {
				radix2Factors[factor][n] = radix2Factors[prev][j]
			}

			for n := 1; n < factor; n += 2 {
				v := -tau / float64(factor) * float64(n)
				radix2Factors[factor][n] = complex(
					trig.Cos(v), trig.Sin(v),
				)
			}
		}
	}

	return radix2Factors[inputLen]
}
