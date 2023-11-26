//nolint:gocognit,lll // skip cyclomatic complexity in benchmarks; comments contain tests output
package osc_test

import (
	"fmt"
	"math"
	"sort"
	"testing"
	"time"

	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/fft"
	"github.com/zalgonoise/x/audio/fft/window"
	"github.com/zalgonoise/x/audio/osc"
)

func TestSawtoothUp(t *testing.T) {
	const (
		sampleRate = 44100
		maxDrift   = 50
		blockSize  = 1024
	)

	// tests that are commented out are currently failing for inaccuracy
	for _, testFreq := range []int{
		13,
		2000,
		3248,
		4000,
		8000,
		16000,
		19983,
	} {
		t.Run(fmt.Sprintf("%dHz", testFreq), func(t *testing.T) {
			// generate wave
			chunk := wav.NewChunk(nil, 16, 1)

			chunk.Generate(osc.SawtoothUpWave, testFreq, sampleRate, 500*time.Millisecond)
			if len(chunk.Value()) == 0 {
				t.Errorf("expected chunk data to be generated")
			}

			// apply FFT to retrieve the frequency spectrum of the signal
			spectrum := fft.Apply(
				sampleRate, chunk.Float()[:blockSize],
				window.New(window.Blackman, blockSize),
			)

			// sort results by highest magnitude first
			sort.Slice(spectrum, func(i, j int) bool {
				return spectrum[i].Mag > spectrum[j].Mag
			})

			// verify that the most powerful frequency is close to the target one
			if spectrum[0].Freq < testFreq-maxDrift ||
				spectrum[0].Freq > testFreq+maxDrift {
				t.Errorf(
					"most powerful frequency is too far off the target: wanted %dHz ; got %dHz",
					testFreq, spectrum[0].Freq,
				)
			}

			t.Logf("got %dHz with magnitude %v", spectrum[0].Freq, spectrum[0].Mag)
		})
	}
}

func TestSawtoothDown(t *testing.T) {
	const (
		sampleRate = 44100
		maxDrift   = 50
		blockSize  = 1024
	)

	// tests that are commented out are currently failing for inaccuracy
	for _, testFreq := range []int{
		13,
		2000,
		3248,
		4000,
		8000,
		16000,
		19983,
	} {
		t.Run(fmt.Sprintf("%dHz", testFreq), func(t *testing.T) {
			// generate wave
			chunk := wav.NewChunk(nil, 16, 1)

			chunk.Generate(osc.SawtoothDownWave, testFreq, sampleRate, 500*time.Millisecond)
			if len(chunk.Value()) == 0 {
				t.Errorf("expected chunk data to be generated")
			}

			// apply FFT to retrieve the frequency spectrum of the signal
			spectrum := fft.Apply(
				sampleRate, chunk.Float()[:blockSize],
				window.New(window.Blackman, blockSize),
			)

			// sort results by highest magnitude first
			sort.Slice(spectrum, func(i, j int) bool {
				return spectrum[i].Mag > spectrum[j].Mag
			})

			// verify that the most powerful frequency is close to the target one
			if spectrum[0].Freq < testFreq-maxDrift ||
				spectrum[0].Freq > testFreq+maxDrift {
				t.Errorf(
					"most powerful frequency is too far off the target: wanted %dHz ; got %dHz",
					testFreq, spectrum[0].Freq,
				)
			}

			t.Logf("got %dHz with magnitude %v", spectrum[0].Freq, spectrum[0].Mag)
		})
	}
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
						var (
							base int16 = ^(2 << int(depth-2)) + 2
							inc        = int16(increment * float64(^base))
						)

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
						chunk = wav.NewChunk(nil, 16, 1)
						chunk.Generate(osc.SawtoothUpWave, 2000, 44100, time.Second/2)
					}

					_ = chunk
				},
			)
			b.Run(
				"ContinuousWrite", func(b *testing.B) {
					chunk := wav.NewChunk(nil, 16, 1)

					for i := 0; i < b.N; i++ {
						chunk.Generate(osc.SawtoothUpWave, 2000, 44100, time.Second/2)
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
						chunk = wav.NewChunk(nil, 16, 1)
						chunk.Generate(osc.SawtoothUpWave, 500, 44100, time.Millisecond*50)
					}

					_ = chunk
				},
			)
			b.Run(
				"ContinuousWrite", func(b *testing.B) {
					chunk := wav.NewChunk(nil, 16, 1)

					for i := 0; i < b.N; i++ {
						chunk.Generate(osc.SawtoothUpWave, 500, 44100, time.Millisecond*50)
					}

					_ = chunk
				},
			)
		},
	)
}
