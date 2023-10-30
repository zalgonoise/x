package trig

import (
	"math"
	"testing"
)

// BenchmarkSin explores performance improvements in this Sin implementation
// and the standar library's math package implementation
//
// goos: linux
// goarch: amd64
// pkg: github.com/zalgonoise/x/audio/wav/fft/trig
// cpu: AMD Ryzen 3 PRO 3300U w/ Radeon Vega Mobile Gfx
// BenchmarkCos/Self-4             539523524               11.12 ns/op            0 B/op          0 allocs/opp
// BenchmarkCos/StdLib-4           504334958               11.61 ns/op            0 B/op          0 allocs/op
// PASS
func BenchmarkCos(b *testing.B) {
	var (
		input   = -1.0
		resultA float64
		resultB float64
	)

	b.Run("Self", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resultA = Cos(input)
		}
		_ = resultA
	})

	b.Run("StdLib", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resultB = math.Cos(input)
		}
		_ = resultB
	})

	if resultA != resultB {
		b.Errorf("output mismatch error: A: %v ; B: %v", resultA, resultB)
	}
}

func FuzzCos(f *testing.F) {
	f.Fuzz(func(t *testing.T, a float64) {
		resA := Cos(a)
		resB := math.Cos(a)
		if resA != resB {
			t.Errorf("output mismatch error: A: %v ; B: %v", resA, resB)
		}
	})
}
