package osc_test

import (
	"math"
	"testing"
	"time"

	"github.com/zalgonoise/x/audio/osc"
	"github.com/zalgonoise/x/audio/wav"
)

func TestSquare(t *testing.T) {
	t.Run(
		"Success", func(t *testing.T) {
			chunk := wav.NewChunk(16, nil)
			chunk.Generate(osc.SquareWave, 2000, 44100, time.Second/2)
			if len(chunk.Value()) == 0 {
				t.Errorf("expected chunk data to be generated")
			}
			t.Logf("%+v", chunk.Value()[:1024])
		},
	)
}

func BenchmarkSquareCompare(b *testing.B) {
	var (
		sampleRate float64 = 44100
		depth      float64 = 16
		freq       float64 = 2000

		before []int16
		after  []int16
	)

	b.Run(
		"squareImplementation", func(b *testing.B) {
			// square function has been improved in performance by calculating the quarter period only once
			// so that the multiplication isn't repeatedly computed; as well as bit-shifting a power of 2
			// instead of calling math.Pow
			//
			// Before / After:
			//
			// goos: linux
			// goarch: amd64
			// cpu: AMD Ryzen 3 PRO 3300U w/ Radeon Vega Mobile Gfx
			// BenchmarkSquareCompare/squareImplementation/square_21_3_23-4             7605760               199.8 ns/op            24 B/op          1 allocs/op
			// BenchmarkSquareCompare/squareImplementation/square_replacement-4         9944721               126.9 ns/op            24 B/op          1 allocs/op
			b.Run(
				"square_21_3_23", func(b *testing.B) {
					fn := func(buffer []int16, halfPeriod int, sampleInt int16) {
						for i := 0; i < len(buffer); i++ {
							if i%halfPeriod < halfPeriod/2 {
								buffer[i] = sampleInt
								continue
							}
							buffer[i] = -sampleInt
						}
					}
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						var (
							halfPeriod = int(sampleRate / (2.0 * freq))
							sampleInt  = int16(math.Pow(2.0, depth-1) - 1.0)
						)
						before = make([]int16, halfPeriod)
						fn(before, halfPeriod, sampleInt)
					}
					_ = before
				},
			)
			b.Run(
				"square_replacement", func(b *testing.B) {
					fn := func(buffer []int16, halfPeriod int, sampleInt int16) {
						// avoid calculating the quarter period over and over again
						var quarterPeriod = halfPeriod / 2
						for i := 0; i < len(buffer); i++ {
							if i%halfPeriod < quarterPeriod {
								buffer[i] = sampleInt
								continue
							}
							buffer[i] = -sampleInt
						}
					}
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						var (
							// avoid floats when calculating halfPeriod
							halfPeriod = int(sampleRate) / (2 * int(freq))
							// bit-shift instead of calculating a power of 2 with math.Pow()
							sampleInt = int16(2<<int16(depth-2)) - 1
						)
						after = make([]int16, halfPeriod)
						fn(after, halfPeriod, sampleInt)
					}
					_ = after
				},
			)
		},
	)
	b.Run(
		"EnsureSimilarResults", func(b *testing.B) {
			if len(before) == 0 || len(after) == 0 {
				b.Error("expected both before and after slices to be populated")
				return
			}

			if len(before) != len(after) {
				b.Errorf("output length mismatch error: wanted %d ; got %d", len(before), len(after))
				return
			}

			for i := range before {
				if before[i] != after[i] {
					b.Errorf("output mismatch error on index #%d -- wanted %d ; got %d", i, before[i], after[i])
					return
				}
			}
		},
	)
}

func BenchmarkSquare(b *testing.B) {
	b.Run(
		"500ms2kHz", func(b *testing.B) {
			b.Run(
				"NilBuffer", func(b *testing.B) {
					var chunk wav.Chunk
					for i := 0; i < b.N; i++ {
						chunk = wav.NewChunk(16, nil)
						chunk.Generate(osc.SquareWave, 2000, 44100, time.Second/2)
					}
					_ = chunk
				},
			)
			b.Run(
				"ContinuousWrite", func(b *testing.B) {
					var chunk = wav.NewChunk(16, nil)
					for i := 0; i < b.N; i++ {
						chunk.Generate(osc.SquareWave, 2000, 44100, time.Second/2)
					}
					_ = chunk
				},
			)
		},
	)
	b.Run(
		"50ms500Hz", func(b *testing.B) {
			b.Run(
				"NilBuffer", func(b *testing.B) {
					var chunk wav.Chunk
					for i := 0; i < b.N; i++ {
						chunk = wav.NewChunk(16, nil)
						chunk.Generate(osc.SquareWave, 500, 44100, time.Millisecond*50)
					}
					_ = chunk
				},
			)
			b.Run(
				"ContinuousWrite", func(b *testing.B) {
					var chunk = wav.NewChunk(16, nil)
					for i := 0; i < b.N; i++ {
						chunk.Generate(osc.SquareWave, 500, 44100, time.Millisecond*50)
					}
					_ = chunk
				},
			)
		},
	)
}