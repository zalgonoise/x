package window

import (
	"math"
	"testing"
)

func BenchmarkHamming(b *testing.B) {
	var (
		outA []float64
		outB []float64
	)

	b.Run("Library", func(b *testing.B) {
		var hamming = func(L int) []float64 {
			r := make([]float64, L)

			if L == 1 {
				r[0] = 1
			} else {
				N := L - 1
				coef := math.Pi * 2 / float64(N)
				for n := 0; n <= N; n++ {
					r[n] = 0.54 - 0.46*math.Cos(coef*float64(n))
				}
			}

			return r
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			outA = hamming(1024)
		}
		_ = outA

	})

	b.Run("Self", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			outB = newHamming(1024)
		}
		_ = outB
	})

	b.Run("Compare", func(b *testing.B) {
		if len(outA) != len(outB) {
			b.Errorf("length mismatch error: wanted %d ; got %d", len(outA), len(outB))
			return
		}
		for i := range outA {
			if outA[i] != outB[i] {
				b.Errorf("output mismatch error: A: %v ; B: %v", outA[i], outB[i])
			}
		}
	})
}

func BenchmarkHann(b *testing.B) {
	var (
		outA []float64
		outB []float64
	)

	b.Run("Library", func(b *testing.B) {
		var hann = func(L int) []float64 {
			r := make([]float64, L)

			if L == 1 {
				r[0] = 1
			} else {
				N := L - 1
				coef := 2 * math.Pi / float64(N)
				for n := 0; n <= N; n++ {
					r[n] = 0.5 * (1 - math.Cos(coef*float64(n)))
				}
			}

			return r
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			outA = hann(1024)
		}
		_ = outA

	})

	b.Run("Self", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			outB = newHann(1024)
		}
		_ = outB
	})

	b.Run("Compare", func(b *testing.B) {
		if len(outA) != len(outB) {
			b.Errorf("length mismatch error: wanted %d ; got %d", len(outA), len(outB))
			return
		}
		for i := range outA {
			if outA[i] != outB[i] {
				b.Errorf("output mismatch error: A: %v ; B: %v", outA[i], outB[i])
			}
		}
	})
}

func BenchmarkBlackman(b *testing.B) {
	var (
		outA []float64
		outB []float64
	)

	b.Run("Library", func(b *testing.B) {
		var blackman = func(L int) []float64 {
			r := make([]float64, L)
			if L == 1 {
				r[0] = 1
			} else {
				N := L - 1
				for n := 0; n <= N; n++ {
					const term0 = 0.42
					term1 := -0.5 * math.Cos(2*math.Pi*float64(n)/float64(N))
					term2 := 0.08 * math.Cos(4*math.Pi*float64(n)/float64(N))
					r[n] = term0 + term1 + term2
				}
			}
			return r
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			outA = blackman(1024)
		}
		_ = outA

	})

	b.Run("Self", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			outB = newBlackman(1024)
		}
		_ = outB
	})

	b.Run("Compare", func(b *testing.B) {
		if len(outA) != len(outB) {
			b.Errorf("length mismatch error: wanted %d ; got %d", len(outA), len(outB))
			return
		}
		for i := range outA {
			if outA[i] != outB[i] {
				b.Errorf("output mismatch error: A: %v ; B: %v", outA[i], outB[i])
			}
		}
	})
}
