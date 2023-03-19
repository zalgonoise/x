package osc_test

import (
	"testing"
	"time"

	"github.com/zalgonoise/x/audio/wav"
	"github.com/zalgonoise/x/audio/wav/data"
	"github.com/zalgonoise/x/audio/wav/osc"
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

func BenchmarkSawtooth(b *testing.B) {
	b.Run(
		"500ms2kHz", func(b *testing.B) {
			b.Run(
				"NilBuffer", func(b *testing.B) {
					var chunk data.Chunk
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
					var chunk data.Chunk
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
