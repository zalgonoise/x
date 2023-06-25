package conv_test

import (
	"reflect"
	"testing"

	"github.com/zalgonoise/x/audio/wav/data"
	"github.com/zalgonoise/x/audio/wav/data/header"
	"github.com/zalgonoise/x/audio/wav/data/internal/testdata/pcm"
)

func BenchmarkChunk16bit(b *testing.B) {
	b.Run(
		"Parse", func(b *testing.B) {
			b.Run(
				"NewBuffer", func(b *testing.B) {
					h, err := header.From(pcm.Test16bitHeader)
					if err != nil {
						b.Error(err)
						return
					}

					var chunk *data.DataChunk
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						chunk = data.NewPCMDataChunk(bitDepth16, h)
						chunk.Parse(pcm.Test16bitPCM)
					}
					_ = chunk
				},
			)
			b.Run(
				"Append", func(b *testing.B) {
					h, err := header.From(pcm.Test16bitHeader)
					if err != nil {
						b.Error(err)
						return
					}

					var chunk = data.NewPCMDataChunk(bitDepth16, h)
					chunk.Parse(pcm.Test16bitPCM)
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						chunk.Parse(pcm.Test16bitPCM)
					}
					_ = chunk
				},
			)
		},
	)
	b.Run(
		"Generate", func(b *testing.B) {
			h, err := header.From(pcm.Test16bitHeader)
			if err != nil {
				b.Error(err)
				return
			}

			var (
				chunk = data.NewPCMDataChunk(bitDepth16, h)
				buf   []byte
			)
			chunk.Parse(pcm.Test16bitPCM)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				buf = chunk.Bytes()
			}
			_ = buf
		},
	)
}

func Test16bitHeader(t *testing.T) {
	h, err := header.From(pcm.Test16bitHeader)
	if err != nil {
		t.Error(err)
		return
	}
	chunk := data.NewPCMDataChunk(bitDepth16, h)

	if output := chunk.Header(); !reflect.DeepEqual(*h, *output) {
		t.Errorf("output mismatch error: wanted %+v ; got %+v", *h, *output)
	}

	if bitDepth := chunk.BitDepth(); bitDepth != chunk.Depth {
		t.Errorf("bit depth mismatch error: wanted %v ; got %v", chunk.Depth, bitDepth)
	}
}
