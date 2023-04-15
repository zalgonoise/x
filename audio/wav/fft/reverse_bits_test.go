package fft

import "testing"

// BenchmarkReverseBits finds the most performant way of computing the reverse bits of
// a number
//
// goos: linux
// goarch: amd64
// pkg: github.com/zalgonoise/x/audio/wav/fft
// cpu: AMD Ryzen 3 PRO 3300U w/ Radeon Vega Mobile Gfx
// BenchmarkReverseBits/Self/NoSize-4         	294989702	         4.224 ns/op	       0 B/op	       0 allocs/op
// BenchmarkReverseBits/Self/WithSize-4       	235033825	         4.652 ns/op	       0 B/op	       0 allocs/op
// BenchmarkReverseBits/GoDSPFFT-4            	218701233	         5.678 ns/op	       0 B/op	       0 allocs/op
// BenchmarkReverseBits/Compare-4             	1000000000	         0.0000015 ns/op	       0 B/op	       0 allocs/op
func BenchmarkReverseBits(b *testing.B) {
	var (
		input  uint = 223
		wants  uint = 251
		amount uint = 8

		outputA uint
		outputB uint
		outputC uint
	)

	b.Run("Self/NoSize", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			outputA = ReverseBits(input)
		}
		_ = outputA
	})
	b.Run("Self/WithSize", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			outputB = ReverseFirstBits(input, amount)
		}
		_ = outputB
	})
	b.Run("GoDSPFFT", func(b *testing.B) {
		var (
			fn = func(v, s uint) uint {
				var r uint

				r = v & 1
				s--

				for v >>= 1; v != 0; v >>= 1 {
					r <<= 1
					r |= v & 1
					s--
				}

				return r << s
			}
		)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			outputC = fn(input, amount)
		}
		_ = outputC
	})
	b.Run("Compare", func(b *testing.B) {
		if outputA != wants || outputB != wants {
			b.Errorf("output mismatch error: A: %d ; B: %d", outputA, outputB)
		}
	})
}
