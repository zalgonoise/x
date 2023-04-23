package osc_test

import (
	"math"
	"testing"
	"time"

	"github.com/zalgonoise/x/audio/osc"
	"github.com/zalgonoise/x/audio/wav"
)

func TestSawtooth(t *testing.T) {
	t.Run(
		"Success", func(t *testing.T) {
			chunk := wav.NewChunk(16, nil)
			chunk.Generate(osc.SawtoothWave, 2000, 44100, time.Second/2)
			if len(chunk.Value()) == 0 {
				t.Errorf("expected chunk data to be generated")
			}
			t.Logf("%+v", chunk.Value()[:1024])
		},
	)
}

func BenchmarkSawtoothCompare(b *testing.B) {
	var (
		sampleRate float64 = 44100
		depth      float64 = 16
		freq       float64 = 2000

		before []int16
		after  []int16
	)

	b.Run(
		"sawtoothImplementation", func(b *testing.B) {
			// sawtooth function has been improved in performance by calculating the base value and increments
			// only once, and by replacing these calculations with only one multiplication
			//
			// Before / After:
			//
			// goos: linux
			// goarch: amd64
			// cpu: AMD Ryzen 3 PRO 3300U w/ Radeon Vega Mobile Gfx
			// BenchmarkSawtoothCompare/sawtoothImplementation/sawtooth_23_3_23-4               1000000              1051 ns/op              48 B/op          1 allocs/op
			// BenchmarkSawtoothCompare/sawtoothImplementation/sawtooth_replacement-4           5365615               231.8 ns/op            48 B/op          1 allocs/op
			b.Run(
				"sawtooth_23_3_23", func(b *testing.B) {
					fn := func(buffer []int16, halfPeriod int, sampleInt int16, increment, depth float64) {
						for i := 0; i < len(buffer); i++ {
							if i%halfPeriod == 0 {
								sampleInt = -int16(math.Pow(2.0, depth-1) - 1.0)
							} else {
								sampleInt += int16(increment * (math.Pow(2.0, depth-1) - 1.0))
							}
							buffer[i] = sampleInt
						}
					}
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						var (
							halfPeriod = int(sampleRate / freq)
							increment  = 2.0 / float64(halfPeriod)
						)
						before = make([]int16, halfPeriod)
						fn(before, halfPeriod, 0, increment, depth)
					}
					_ = before
				},
			)
			b.Run(
				"sawtooth_replacement", func(b *testing.B) {
					fn := func(buffer []int16, halfPeriod int, sampleInt int16, increment, depth float64) {
						var base int16 = ^(2 << int(depth-2)) + 2
						inc := int16(increment * float64(^base))

						for i := 0; i < len(buffer); i++ {
							if i%halfPeriod == 0 {
								sampleInt = base
							} else {
								sampleInt += inc
							}
							buffer[i] = sampleInt
						}
					}
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						var (
							halfPeriod = int(sampleRate / freq)
							increment  = 2.0 / float64(halfPeriod)
						)
						after = make([]int16, halfPeriod)
						fn(after, halfPeriod, 0, increment, depth)
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

func BenchmarkSawtooth(b *testing.B) {
	b.Run(
		"500ms2kHz", func(b *testing.B) {
			b.Run(
				"NilBuffer", func(b *testing.B) {
					var chunk wav.Chunk
					for i := 0; i < b.N; i++ {
						chunk = wav.NewChunk(16, nil)
						chunk.Generate(osc.SawtoothWave, 2000, 44100, time.Second/2)
					}
					_ = chunk
				},
			)
			b.Run(
				"ContinuousWrite", func(b *testing.B) {
					var chunk = wav.NewChunk(16, nil)
					for i := 0; i < b.N; i++ {
						chunk.Generate(osc.SawtoothWave, 2000, 44100, time.Second/2)
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
						chunk.Generate(osc.SawtoothWave, 500, 44100, time.Millisecond*50)
					}
					_ = chunk
				},
			)
			b.Run(
				"ContinuousWrite", func(b *testing.B) {
					var chunk = wav.NewChunk(16, nil)
					for i := 0; i < b.N; i++ {
						chunk.Generate(osc.SawtoothWave, 500, 44100, time.Millisecond*50)
					}
					_ = chunk
				},
			)
		},
	)
}
