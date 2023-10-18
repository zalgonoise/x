package conv_test

import (
	"reflect"
	"testing"

	"github.com/zalgonoise/x/audio/encoding/wav/data"
	"github.com/zalgonoise/x/audio/encoding/wav/data/header"
	"github.com/zalgonoise/x/audio/encoding/wav/data/internal/testdata/pcm"
)

const (
	bitDepth8  uint16 = 8
	bitDepth16 uint16 = 16
	bitDepth24 uint16 = 24
	bitDepth32 uint16 = 32
)

func BenchmarkChunk8bit(b *testing.B) {
	b.Run(
		"Parse", func(b *testing.B) {
			b.Run(
				"NewBuffer", func(b *testing.B) {
					h, err := header.From(pcm.Test8bitHeader)
					if err != nil {
						b.Error(err)
						return
					}

					var chunk *data.DataChunk
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						chunk = data.NewPCMDataChunk(bitDepth8, h)
						chunk.Parse(pcm.Test8bitPCM)
					}
					_ = chunk
				},
			)
			b.Run(
				"Append", func(b *testing.B) {
					h, err := header.From(pcm.Test8bitHeader)
					if err != nil {
						b.Error(err)
						return
					}

					var chunk = data.NewPCMDataChunk(bitDepth8, h)
					chunk.Parse(pcm.Test8bitPCM)
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						chunk.Parse(pcm.Test8bitPCM)
					}
					_ = chunk
				},
			)
		},
	)
	b.Run(
		"Bytes", func(b *testing.B) {
			h, err := header.From(pcm.Test8bitHeader)
			if err != nil {
				b.Error(err)
				return
			}

			var (
				chunk = data.NewPCMDataChunk(bitDepth8, h)
				buf   []byte
			)
			chunk.Parse(pcm.Test8bitPCM)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				buf = chunk.Bytes()
			}
			_ = buf
		},
	)
}

func Test8bitHeader(t *testing.T) {
	h, err := header.From(pcm.Test8bitHeader)
	if err != nil {
		t.Error(err)
		return
	}
	chunk := &data.DataChunk{
		ChunkHeader: h,
		Depth:       8, // set by NewChunk()
	}

	if output := chunk.Header(); !reflect.DeepEqual(*h, *output) {
		t.Errorf("output mismatch error: wanted %+v ; got %+v", *h, *output)
	}

	if bitDepth := chunk.BitDepth(); bitDepth != chunk.Depth {
		t.Errorf("bit depth mismatch error: wanted %v ; got %v", chunk.Depth, bitDepth)
	}
}
