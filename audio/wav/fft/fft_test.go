package fft

import (
	"math"
	"testing"
)

// BenchmarkHypotenuse compares three approaches to calculating the hypotenuse using Go's standard library,
// to prioritize the most performant approach -- that is, by declaring variables for the real and imaginary parts of the
// complex number in question, and calculating `sqrt(real*real + imag*imag)`
//
// ❯ go test -bench . -benchmem -benchtime=5s -cpuprofile /tmp/cpu.pprof ./fft
//
// goos: linux
// goarch: amd64
// pkg: github.com/zalgonoise/x/audio/wav/fft
// cpu: AMD Ryzen 3 PRO 3300U w/ Radeon Vega Mobile Gfx
// BenchmarkHypotenuse/Simplified-4                60037294                18.63 ns/op            0 B/op          0 allocs/op
// BenchmarkHypotenuse/Minimal-4                   415254681                3.015 ns/op           0 B/op          0 allocs/op
// BenchmarkHypotenuse/Extended-4                  330081352                3.616 ns/op           0 B/op          0 allocs/op
func BenchmarkHypotenuse(b *testing.B) {
	var complexData = []complex128{
		0.5 + 1.3i, 1.1 + 0.4i, 0.8 - 1.2i, 1.3 - 0.5i,
	}

	b.Run("Simplified", func(b *testing.B) {
		var out [4]float64
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for idx := range complexData {
				out[idx] = math.Hypot(real(complexData[idx]), imag(complexData[idx]))
			}
		}
		_ = out
	})

	b.Run("Minimal", func(b *testing.B) {
		var out [4]float64
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for idx := range complexData {
				re := real(complexData[idx])
				im := imag(complexData[idx])
				out[idx] = math.Sqrt(re*re + im*im)
			}
		}
		_ = out
	})

	b.Run("Extended", func(b *testing.B) {
		var out [4]float64
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for idx := range complexData {
				out[idx] = math.Sqrt(
					real(complexData[idx])*real(complexData[idx]) +
						imag(complexData[idx])*imag(complexData[idx]),
				)
			}
		}
		_ = out
	})

}
