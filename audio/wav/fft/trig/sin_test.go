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
// pkg: github.com/zalgonoise/x/audio/wav/fft
// cpu: AMD Ryzen 3 PRO 3300U w/ Radeon Vega Mobile Gfx
// BenchmarkSin/Self-4         	83990611	        14.14 ns/op	       0 B/op	       0 allocs/op
// BenchmarkSin/StdLib-4       	75234088	        15.58 ns/op	       0 B/op	       0 allocs/op
// PASS
func BenchmarkSin(b *testing.B) {
	var (
		input   = -1.0
		resultA float64
		resultB float64
	)

	b.Run("Self", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resultA = Sin(input)
		}
		_ = resultA
	})

	b.Run("StdLib", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resultB = math.Sin(input)
		}
		_ = resultB
	})

	if resultA != resultB {
		b.Errorf("output mismatch error: A: %v ; B: %v", resultA, resultB)
	}
}

func FuzzSin(f *testing.F) {

	f.Add(1.2)
	f.Add(0.3)
	f.Add(8.0)
	f.Add(123.7)
	f.Add(-1.0)
	f.Add(-0.3)
	f.Add(-5.0)
	f.Add(536870913.0)
	f.Fuzz(func(t *testing.T, a float64) {
		resA := Sin(a)
		resB := math.Sin(a)
		if resA != resB {
			t.Errorf("output mismatch error: A: %v ; B: %v", resA, resB)
		}
	})
}
