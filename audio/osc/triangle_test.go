package osc_test

import (
	"math"
	"testing"
	"time"

	"github.com/zalgonoise/x/audio/osc"
	"github.com/zalgonoise/x/audio/wav"
)

func TestTriangle(t *testing.T) {
	t.Run(
		"Success", func(t *testing.T) {
			chunk := wav.NewChunk(16, nil)
			chunk.Generate(osc.TriangleWave, 2000, 44100, time.Millisecond*500)
			if len(chunk.Value()) == 0 {
				t.Errorf("expected chunk data to be generated")
			}
			t.Logf("%+v", chunk.Value()[:1024])
		},
	)
}

func BenchmarkTriangleCompare(b *testing.B) {
	var (
		sampleRate float64 = 44100
		depth      float64 = 16
		freq       float64 = 2000

		before []int16
		after  []int16
	)

	b.Run(
		"triangleImplementation", func(b *testing.B) {
			// triangle function has been improved in performance by calculating the base value and increments
			// only once, and by replacing these calculations with only one multiplication. Calls to math.Pow()
			// have been replaced with bit-shifting. The quarter period value is also precalculated
			//
			// Before / After:
			//
			// goos: linux
			// goarch: amd64
			// cpu: AMD Ryzen 3 PRO 3300U w/ Radeon Vega Mobile Gfx
			// BenchmarkTriangleCompare/triangleImplementation/triangle_24_3_23-4               1000000              1066 ns/op              48 B/op          1 allocs/op
			// BenchmarkTriangleCompare/triangleImplementation/triangle_replacement-4           4885150               210.8 ns/op            48 B/op          1 allocs/op
			b.Run(
				"triangle_24_3_23", func(b *testing.B) {
					fn := func(buffer []int16, halfPeriod int, sampleInt int16, increment, depth float64) {
						var swap bool
						for i := 0; i < len(buffer); i++ {
							if i%(halfPeriod/2) == 0 {
								swap = !swap
							}
							if swap {
								sampleInt += int16(increment * (math.Pow(2.0, depth-1) - 1.0))
							} else {
								sampleInt -= int16(increment * (math.Pow(2.0, depth-1) - 1.0))
							}
							buffer[i] = sampleInt
						}
					}
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						var (
							halfPeriod       = int(sampleRate / freq)
							increment        = 4.0 / float64(halfPeriod)
							sampleInt  int16 = -(1 << int(depth-1))
						)
						before = make([]int16, halfPeriod)
						fn(before, halfPeriod, sampleInt, increment, depth)
					}
					_ = before
				},
			)
			b.Run(
				"triangle_replacement", func(b *testing.B) {
					fn := func(buffer []int16, halfPeriod int, sampleInt int16, increment, depth float64) {
						var stepValue = int16(increment * float64(int(2)<<int(depth-2)-1))
						var quarterPeriod = halfPeriod / 2
						var swap bool
						for i := 0; i < len(buffer); i++ {
							if i%(quarterPeriod) == 0 {
								swap = !swap
							}
							if swap {
								sampleInt += stepValue
							} else {
								sampleInt -= stepValue
							}
							buffer[i] = sampleInt
						}
					}
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						var (
							halfPeriod       = int(sampleRate / freq)
							increment        = 4.0 / float64(halfPeriod)
							sampleInt  int16 = -(1 << int(depth-1))
						)
						after = make([]int16, halfPeriod)
						fn(after, halfPeriod, sampleInt, increment, depth)
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

func BenchmarkTriangle(b *testing.B) {
	b.Run(
		"500ms2kHz", func(b *testing.B) {
			b.Run(
				"NilBuffer", func(b *testing.B) {
					var chunk wav.Chunk
					for i := 0; i < b.N; i++ {
						chunk = wav.NewChunk(16, nil)
						chunk.Generate(osc.TriangleWave, 2000, 44100, time.Second/2)
					}
					_ = chunk
				},
			)
			b.Run(
				"ContinuousWrite", func(b *testing.B) {
					var chunk = wav.NewChunk(16, nil)
					for i := 0; i < b.N; i++ {
						chunk.Generate(osc.TriangleWave, 2000, 44100, time.Second/2)
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
						chunk.Generate(osc.TriangleWave, 500, 44100, time.Millisecond*50)
					}
					_ = chunk
				},
			)
			b.Run(
				"ContinuousWrite", func(b *testing.B) {
					var chunk = wav.NewChunk(16, nil)
					for i := 0; i < b.N; i++ {
						chunk.Generate(osc.TriangleWave, 500, 44100, time.Millisecond*50)
					}
					_ = chunk
				},
			)
		},
	)
}