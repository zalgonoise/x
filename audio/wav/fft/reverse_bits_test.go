package fft_test

import (
	"testing"

	"github.com/zalgonoise/x/audio/wav/fft"
)

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
			outputA = fft.ReverseBits(input)
		}
		_ = outputA
	})
	b.Run("Self/WithSize", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			outputB = fft.ReverseFirstBits(input, amount)
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

func BenchmarkReorderData(b *testing.B) {
	sine, err := newSine(2000)
	if err != nil {
		b.Error(err)
		return
	}

	var (
		data      = fft.ToComplex(sine.Data.Float())[:1024]
		spectrumA []complex128
		spectrumB []complex128
	)
	b.Run("Self/ReorderData", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			spectrumA = fft.ReorderData(data)
		}
		_ = spectrumA
	})

	b.Run("GoDSP/ReorderData", func(b *testing.B) {
		var fn = func(x []complex128) []complex128 {
			ln := uint(len(x))
			reorder := make([]complex128, ln)
			s := fft.Log2(ln)

			var n uint
			for ; n < ln; n++ {
				reorder[fft.ReverseFirstBits(n, s)] = x[n]
			}

			return reorder
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			spectrumB = fn(data)
		}
		_ = spectrumB
	})
}

func TestReorderData(t *testing.T) {
	sine, err := newSine(2000)
	if err != nil {
		t.Error(err)
		return
	}

	var (
		data      = fft.ToComplex(sine.Data.Float())[:1024]
		spectrumA []complex128
		spectrumB []complex128
	)
	t.Run("Self/ReorderData", func(t *testing.T) {
		spectrumA = fft.ReorderData(data)
	})

	t.Run("GoDSP/ReorderData", func(t *testing.T) {
		spectrumB = func(x []complex128) []complex128 {
			ln := uint(len(x))
			reorder := make([]complex128, ln)
			s := fft.Log2(ln)

			var n uint
			for ; n < ln; n++ {
				reorder[fft.ReverseFirstBits(n, s)] = x[n]
			}

			return reorder
		}(data)
	})

	t.Run("Compare", func(t *testing.T) {
		if len(spectrumA) != len(spectrumB) {
			t.Errorf("output length mismatch error: slice A: %d ; slice B: %d", len(spectrumA), len(spectrumB))
			return
		}

		for idx := range spectrumA {
			if spectrumA[idx] != spectrumB[idx] {
				t.Errorf("output mismatch error: index #%d: slice A: %v ; slice B: %v", idx, spectrumA[idx], spectrumB[idx])
			}
		}
	})
}
