package conv_test

import (
	"github.com/zalgonoise/x/audio/wav/data"
	"github.com/zalgonoise/x/audio/wav/data/internal/testdata"
	"reflect"
	"testing"

	"github.com/zalgonoise/x/audio/wav/data/header"
)

func BenchmarkChunk24bit(b *testing.B) {
	b.Run(
		"Parse", func(b *testing.B) {
			b.Run(
				"NewBuffer", func(b *testing.B) {
					h, err := header.From(testdata.Test24bitHeader)
					if err != nil {
						b.Error(err)
						return
					}

					var chunk *data.DataChunk
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						chunk = data.NewPCMDataChunk(bitDepth24, h)
						chunk.Parse(testdata.Test24bitPCM)
					}
					_ = chunk
				},
			)
			b.Run(
				"Append", func(b *testing.B) {
					h, err := header.From(testdata.Test24bitHeader)
					if err != nil {
						b.Error(err)
						return
					}

					var chunk = data.NewPCMDataChunk(bitDepth24, h)
					chunk.Parse(testdata.Test24bitPCM)
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						chunk.Parse(testdata.Test24bitPCM)
					}
					_ = chunk
				},
			)
		},
	)
	b.Run(
		"Generate", func(b *testing.B) {
			h, err := header.From(testdata.Test24bitHeader)
			if err != nil {
				b.Error(err)
				return
			}

			var (
				chunk = data.NewPCMDataChunk(bitDepth24, h)
				buf   []byte
			)
			chunk.Parse(testdata.Test24bitPCM)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				buf = chunk.Bytes()
			}
			_ = buf
		},
	)
}

func Test24bitHeader(t *testing.T) {
	h, err := header.From(testdata.Test24bitHeader)
	if err != nil {
		t.Error(err)
		return
	}
	chunk := data.NewPCMDataChunk(bitDepth24, h)

	if output := chunk.Header(); !reflect.DeepEqual(*h, *output) {
		t.Errorf("output mismatch error: wanted %+v ; got %+v", *h, *output)
	}

	if bitDepth := chunk.BitDepth(); bitDepth != chunk.Depth {
		t.Errorf("bit depth mismatch error: wanted %v ; got %v", chunk.Depth, bitDepth)
	}
}
