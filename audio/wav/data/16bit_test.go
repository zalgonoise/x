package data

import (
	"bytes"
	"reflect"
	"testing"
)

var (
	test16bitPCM = []byte{
		0x32, 0x9, 0x87, 0xfc, 0xdc, 0xf4, 0x5f, 0xf6, 0x76, 0x1, 0x28, 0xf, 0xb, 0x16, 0xb0, 0x15, 0xcb, 0xd, 0x80, 0x0,
		0xac, 0xfd, 0xe5, 0xa, 0x95, 0x16, 0x54, 0x16, 0x49, 0xf, 0x5f, 0x6, 0x48, 0x3, 0x6b, 0xc, 0x90, 0x1a, 0xce, 0x23,
		0x33, 0x23, 0x14, 0x14, 0x9d, 0xfc, 0x2f, 0xef, 0x86, 0xf3, 0x93, 0x1, 0xe3, 0x10, 0xa9, 0x1c, 0x4f, 0x21, 0x55,
		0x1d, 0xeb, 0x10, 0xa8, 0x2, 0x9c, 0x0, 0x74, 0xd, 0x60, 0x16, 0x43, 0xc, 0x7d, 0xf9, 0xa0, 0xf2, 0x49, 0xfe, 0x13,
		0x12, 0xff, 0x1e, 0xb, 0x1c, 0x1a, 0x9, 0x5, 0xf5, 0xb1, 0xf9, 0x44, 0x17, 0x96, 0x28, 0x2a, 0x1a, 0x80, 0x0, 0x7f,
		0xeb, 0x75, 0xe0, 0x2f, 0xec, 0x9d, 0x8, 0x14, 0x21, 0x8a, 0x35, 0x9a, 0x3e, 0x3c, 0x21, 0xc6, 0xea, 0xc, 0xcb,
		0x14, 0xcd, 0xad, 0xdc, 0x3f, 0xf3, 0x17, 0xf, 0xa, 0x24,
	}
	test16bitHeader = []byte{0x64, 0x61, 0x74, 0x61, 0x8, 0x49, 0x0, 0x0}
)

func BenchmarkChunk16bit(b *testing.B) {
	b.Run(
		"Parse", func(b *testing.B) {
			b.Run(
				"NewBuffer", func(b *testing.B) {
					header, err := HeaderFrom(test16bitHeader)
					if err != nil {
						b.Error(err)
						return
					}

					var chunk *Chunk16bit
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						chunk = &Chunk16bit{
							ChunkHeader: header,
							Depth:       16, // set by NewChunk()
						}
						chunk.Parse(test16bitPCM)
					}
					_ = chunk
				},
			)
			b.Run(
				"Append", func(b *testing.B) {
					header, err := HeaderFrom(test16bitHeader)
					if err != nil {
						b.Error(err)
						return
					}

					var chunk = &Chunk16bit{
						ChunkHeader: header,
						Depth:       16, // set by NewChunk()
					}
					chunk.Parse(test16bitPCM)
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						chunk.Parse(test16bitPCM)
					}
					_ = chunk
				},
			)
		},
	)
	b.Run(
		"Generate", func(b *testing.B) {
			header, err := HeaderFrom(test16bitHeader)
			if err != nil {
				b.Error(err)
				return
			}

			var (
				chunk = &Chunk16bit{
					ChunkHeader: header,
					Depth:       16, // set by NewChunk()
				}
				buf []byte
			)
			chunk.Parse(test16bitPCM)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				buf = chunk.Bytes()
			}
			_ = buf
		},
	)
}

func Test16bitHeader(t *testing.T) {
	header, err := HeaderFrom(test16bitHeader)
	if err != nil {
		t.Error(err)
		return
	}
	chunk := &Chunk16bit{
		ChunkHeader: header,
		Depth:       16, // set by NewChunk()
	}

	if output := chunk.Header(); !reflect.DeepEqual(*header, *output) {
		t.Errorf("output mismatch error: wanted %+v ; got %+v", *header, *output)
	}

	if bitDepth := chunk.BitDepth(); bitDepth != chunk.Depth {
		t.Errorf("bit depth mismatch error: wanted %v ; got %v", chunk.Depth, bitDepth)
	}
}

func Test16bitParse(t *testing.T) {
	header, err := HeaderFrom(test16bitHeader)
	if err != nil {
		t.Error(err)
		return
	}
	chunk := &Chunk16bit{
		ChunkHeader: header,
	}

	chunk.Parse(test16bitPCM)
	buf := chunk.Bytes()
	if !bytes.Equal(test16bitPCM, buf) {
		t.Errorf("output mismatch error: input is not the same as output")
	}

	if i := chunk.Value(); len(i) == 0 {
		t.Errorf("expected integer PCM buffer to be longer than zero")
	}

	if f := chunk.Float(); len(f) == 0 {
		t.Errorf("expected float PCM buffer to be longer than zero")
	}

	if chunk.Reset(); chunk.Data != nil {
		t.Errorf("expected Reset() method to clear the data in the chunk")
	}
}