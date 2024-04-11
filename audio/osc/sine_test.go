//nolint:lll // comments contain tests output
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
	"github.com/zalgonoise/x/audio/trig"
)

func TestSine(t *testing.T) {
	const (
		sampleRate = 44100
		maxDrift   = 50
		blockSize  = 1024
	)

	for _, testFreq := range []int{
		1,
		13,
		2000,
		4000,
		8000,
		16000,
		19983,
		22000,
	} {
		t.Run(fmt.Sprintf("%dHz", testFreq), func(t *testing.T) {
			// generate wave
			chunk := wav.NewChunk(nil, 16, 1)
			chunk.Generate(osc.SineWave, testFreq, sampleRate, 750*time.Millisecond)

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

			t.Log(chunk.Float()[:128])
			t.Logf("got %dHz with magnitude %v", spectrum[0].Freq, spectrum[0].Mag)
		})
	}
}

func BenchmarkSineCompare(b *testing.B) {
	var (
		sampleRate float64 = 44100
		depth      float64 = 16
		freq       float64 = 2000
		halfPeriod         = int(sampleRate / freq)

		before []int16
		after  []int16
	)

	const (
		tau float64 = math.Pi * 2
	)

	b.Run(
		"sineImplementation", func(b *testing.B) {
			// sine function has been improved in performance by setting Tau (2Pi) as a constant
			// so that the multiplication isn't repeatedly computed; as well as bit-shifting a power of 2
			// instead of calling math.Pow
			//
			// Before / After:
			//
			// goos: linux
			// goarch: amd64
			// cpu: AMD Ryzen 3 PRO 3300U w/ Radeon Vega Mobile Gfx
			// BenchmarkSineCompare/sineImplementation/sine_20_3_23-4           1281918              1969 ns/op              48 B/op          1 allocs/op
			// BenchmarkSineCompare/sineImplementation/sine_replacement-4       2662305               556.2 ns/op            48 B/op          1 allocs/op
			b.Run(
				"sine_20_3_23", func(b *testing.B) {
					fn := func(buffer []int16, freq, depth, sampleRate float64) {
						for i := 0; i < len(buffer); i++ {
							sample := math.Sin(2.0 * math.Pi * freq * float64(i) / sampleRate)
							buffer[i] = int16(sample * (math.Pow(2.0, depth)/2.0 - 1.0))
						}
					}

					b.ResetTimer()

					for i := 0; i < b.N; i++ {
						before = make([]int16, halfPeriod)
						fn(before, freq, depth, sampleRate)
					}

					_ = before
				},
			)
			b.Run(
				"sine_replacement", func(b *testing.B) {
					// tau is now a constant in the package
					fn := func(buffer []int16, freq, depth, sampleRate float64) {
						for i := 0; i < len(buffer); i++ {
							// use tau here, no need to compute 2 * math.Pi
							sample := trig.Sin(tau * freq * float64(i) / sampleRate)
							// bit shift here, instead of math.Pow(2.0, depth); which is equivalent to 2 << (depth-1)
							buffer[i] = int16(sample * float64(int(2)<<int(depth-1)/2-1))
						}
					}

					b.ResetTimer()

					for i := 0; i < b.N; i++ {
						after = make([]int16, halfPeriod)
						fn(after, freq, depth, sampleRate)
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

func BenchmarkSine(b *testing.B) {
	b.Run(
		"500ms2kHz", func(b *testing.B) {
			b.Run(
				"NilBuffer", func(b *testing.B) {
					var chunk wav.Chunk

					for i := 0; i < b.N; i++ {
						chunk = wav.NewChunk(nil, 16, 1)
						chunk.Generate(osc.SineWave, 2000, 44100, time.Second/2)
					}

					_ = chunk
				},
			)
			b.Run(
				"ContinuousWrite", func(b *testing.B) {
					chunk := wav.NewChunk(nil, 16, 1)

					for i := 0; i < b.N; i++ {
						chunk.Generate(osc.SineWave, 2000, 44100, time.Second/2)
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
						chunk.Generate(osc.SineWave, 500, 44100, time.Millisecond*50)
					}

					_ = chunk
				},
			)
			b.Run(
				"ContinuousWrite", func(b *testing.B) {
					chunk := wav.NewChunk(nil, 16, 1)

					for i := 0; i < b.N; i++ {
						chunk.Generate(osc.SineWave, 500, 44100, time.Millisecond*50)
					}

					_ = chunk
				},
			)
		},
	)
}
