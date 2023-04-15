package fft_test

import (
	"math"
	"sync"
	"testing"

	"github.com/zalgonoise/x/audio/wav/fft"
)

// BenchmarkGetRadix2Factors tests the current and former implementations of
// this function, to measure its improvements over time
//
// ❯ go test -v  -bench '^(BenchmarkGetRadix2Factors)$' -run='^$'  -benchmem -benchtime=1s -cpuprofile /tmp/cpu.pprof ./wav/fft
//
// goos: linux
// goarch: amd64
// pkg: github.com/zalgonoise/x/audio/wav/fft
// cpu: AMD Ryzen 3 PRO 3300U w/ Radeon Vega Mobile Gfx
// BenchmarkGetRadix2Factors
// BenchmarkGetRadix2Factors/Self/GetRadix2Factors/Improvement_13_04_2023
// BenchmarkGetRadix2Factors/Self/GetRadix2Factors/Improvement_13_04_2023-4                100000000               10.46 ns/op            0 B/op          0 allocs/op
// BenchmarkGetRadix2Factors/Self/GetRadix2Factors/Initial
// BenchmarkGetRadix2Factors/Self/GetRadix2Factors/Initial-4                               112066077               14.25 ns/op            0 B/op          0 allocs/op
// BenchmarkGetRadix2Factors/Self/GetRadix2Factors/Original
// BenchmarkGetRadix2Factors/Self/GetRadix2Factors/Original-4                              39731154                30.11 ns/op            0 B/op          0 allocs/op
// BenchmarkGetRadix2Factors/Compare
// BenchmarkGetRadix2Factors/Compare-4
func BenchmarkGetRadix2Factors(b *testing.B) {
	const (
		tau   = 2 * math.Pi
		input = 8192
	)

	var (
		outputA []complex128
		outputB []complex128
		outputC []complex128
	)

	b.Run("Self/GetRadix2Factors/Improvement_13_04_2023", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			outputA = fft.GetRadix2Factors(input)
		}
		_ = outputA
	})

	b.Run("Self/GetRadix2Factors/Initial", func(b *testing.B) {
		var (
			radix2Factors = map[int][]complex128{
				4: {(1 + 0i), (0 - 1i), (-1 + 0i), (0 + 1i)},
			}
			fn = func(inputLen int) []complex128 {
				if factors, ok := radix2Factors[inputLen]; ok {
					return factors
				}

				for factor, prev := 8, 4; factor <= inputLen; factor, prev = factor<<1, factor {
					if _, ok := radix2Factors[factor]; !ok {
						radix2Factors[factor] = make([]complex128, factor)

						for n, j := 0, 0; n < factor; n, j = n+2, j+1 {
							radix2Factors[factor][n] = radix2Factors[prev][j]
						}

						for n := 1; n < factor; n += 2 {
							v := -tau / float64(factor) * float64(n)
							sin, cos := math.Sin(v), math.Cos(v)
							radix2Factors[factor][n] = complex(cos, sin)
						}
					}
				}

				return radix2Factors[inputLen]
			}
		)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			outputB = fn(input)
		}
		_ = outputB
	})
	b.Run("Self/GetRadix2Factors/Original", func(b *testing.B) {
		var (
			radix2Factors = map[int][]complex128{
				4: {(1 + 0i), (0 - 1i), (-1 + 0i), (0 + 1i)},
			}
			radix2Lock       sync.RWMutex
			hasRadix2Factors = func(idx int) bool {
				return radix2Factors[idx] != nil
			}
			fn = func(input_len int) []complex128 {
				radix2Lock.RLock()

				if hasRadix2Factors(input_len) {
					defer radix2Lock.RUnlock()
					return radix2Factors[input_len]
				}

				radix2Lock.RUnlock()
				radix2Lock.Lock()
				defer radix2Lock.Unlock()

				if !hasRadix2Factors(input_len) {
					for i, p := 8, 4; i <= input_len; i, p = i<<1, i {
						if radix2Factors[i] == nil {
							radix2Factors[i] = make([]complex128, i)

							for n, j := 0, 0; n < i; n, j = n+2, j+1 {
								radix2Factors[i][n] = radix2Factors[p][j]
							}

							for n := 1; n < i; n += 2 {
								sin, cos := math.Sincos(-2 * math.Pi / float64(i) * float64(n))
								radix2Factors[i][n] = complex(cos, sin)
							}
						}
					}
				}

				return radix2Factors[input_len]
			}
		)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			outputC = fn(input)
		}
		_ = outputC
	})

	b.Run("Compare", func(b *testing.B) {
		if len(outputA) != len(outputB) || len(outputA) != len(outputC) {
			b.Errorf("output length mismatch error: slice A: %d ; slice B: %d ; slice C: %d", len(outputA), len(outputB), len(outputC))
			return
		}

		for idx := range outputA {
			if outputA[idx] != outputB[idx] || outputA[idx] != outputC[idx] {
				b.Errorf("output mismatch error: index #%d: slice A: %v ; slice B: %v ; slice C: %v", idx, outputA[idx], outputB[idx], outputC[idx])
				// return
			}
		}
	})

}

func TestGetRadix2Factors(t *testing.T) {
	var inputFactorLengths = []int{4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192}

	for _, v := range inputFactorLengths {
		t.Logf("\t%d: %#v,\n", v, fft.GetRadix2Factors(v))
	}
}
