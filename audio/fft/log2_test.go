//nolint:cyclop,godot // tests and benchmarks are excluded from cyclomatic complexity checks
package fft

import "testing"

// BenchmarkLog2 finds the most performant way of computing the log base 2 of
// a number
//
// goos: linux
// goarch: amd64
// pkg: github.com/zalgonoise/x/audio/wav/fft
// cpu: AMD Ryzen 3 PRO 3300U w/ Radeon Vega Mobile Gfx
// BenchmarkLog2
// BenchmarkLog2/HardcodedSwitch
// BenchmarkLog2/HardcodedSwitch-4                 413181310                2.892 ns/op           0 B/op          0 allocs/op
// BenchmarkLog2/HardcodedTable
// BenchmarkLog2/HardcodedTable-4                  157142354                7.988 ns/op           0 B/op          0 allocs/op
// BenchmarkLog2/Self
// BenchmarkLog2/Self-4                            208779463                5.473 ns/op           0 B/op          0 allocs/op
// BenchmarkLog2/GoDSPFFT
// BenchmarkLog2/GoDSPFFT-4                        181338795                6.071 ns/op           0 B/op          0 allocs/op
// BenchmarkLog2/Compare
// BenchmarkLog2/Compare-4                         1000000000               0.0000020 ns/op               0 B/op          0 allocs/op
// PASS
func BenchmarkLog2(b *testing.B) {
	var (
		fullInput = []uint{4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192}
		input     = fullInput[11] // 8192

		outputA uint
		outputB uint
		outputC uint
		outputD uint
	)

	b.Run("HardcodedSwitch", func(b *testing.B) {
		log2fn := func(v uint) uint {
			var r uint
			for ; v > 1; v >>= 1 {
				r++
			}

			return r
		}

		fn := func(v uint) uint {
			switch v {
			case 2:
				return 1
			case 4:
				return 2
			case 8:
				return 3
			case 16:
				return 4
			case 32:
				return 5
			case 64:
				return 6
			case 128:
				return 7
			case 256:
				return 8
			case 512:
				return 9
			case 1024:
				return 10
			case 2048:
				return 11
			case 4096:
				return 12
			case 8192:
				return 13
			default:
				return log2fn(v)
			}
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			outputA = fn(input)
		}

		_ = outputA
	})

	b.Run("HardcodedTable", func(b *testing.B) {
		log2Table := map[uint]uint{
			4:    2,
			8:    3,
			16:   4,
			32:   5,
			64:   6,
			128:  7,
			256:  8,
			512:  9,
			1024: 10,
			2048: 11,
			4096: 12,
			8192: 13,
		}

		for i := 0; i < b.N; i++ {
			outputB = log2Table[input]
		}

		_ = outputB
	})

	b.Run("Self", func(b *testing.B) {
		fn := func(v uint) uint {
			var r uint
			for ; v > 1; v >>= 1 {
				r++
			}

			return r
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			outputC = fn(input)
		}

		_ = outputC
	})

	b.Run("GoDSPFFT", func(b *testing.B) {
		fn := func(v uint) uint {
			var r uint

			for v >>= 1; v != 0; v >>= 1 {
				r++
			}

			return r
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			outputD = fn(input)
		}

		_ = outputD
	})

	b.Run("Compare", func(b *testing.B) {
		if outputA != outputB || outputA != outputC || outputA != outputD {
			b.Errorf("output mismatch error: input: %d;  A: %d ; B: %d ; C: %d ; D: %d", input, outputA, outputB, outputC, outputD)
		}
	})
}
