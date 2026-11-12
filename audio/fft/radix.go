//nolint:mnd // contains hardcoded constants, it's less readable to make constants from them
package fft

import (
	"slices"
	"sync"

	"github.com/zalgonoise/x/audio/trig"
)

const (
	firstCur  = 2048
	firstPrev = 1024
)

// radix2Mu guards radix2Factors, which FFT fills lazily for sizes beyond the precomputed ones: concurrent FFTs of a
// new size would otherwise write the map concurrently (a fatal error in Go).
var radix2Mu sync.RWMutex //nolint:gochecknoglobals // guards the package-level twiddle cache

// GetRadix2Factors returns the radix-2 twiddle factors for an FFT of size inputLen (a power of two).
//
// It returns a copy, so callers cannot corrupt the cache used by FFT. It is temporarily public, and could become
// private at a later point.
func GetRadix2Factors(inputLen int) []complex128 {
	return slices.Clone(radix2(inputLen))
}

// radix2 returns the cached twiddle factors for an FFT of size inputLen, computing and caching any missing sizes.
// The returned slice is shared and must not be modified.
func radix2(inputLen int) []complex128 {
	radix2Mu.RLock()

	factors, ok := radix2Factors[inputLen]

	radix2Mu.RUnlock()

	if ok {
		return factors
	}

	radix2Mu.Lock()
	defer radix2Mu.Unlock()

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
