package data

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/zalgonoise/x/audio/osc"
)

var (
	test32bitPCM = []byte{
		0x0, 0xab, 0x31, 0x9, 0x0, 0xfb, 0x86, 0xfc, 0x0, 0x6f, 0xdc, 0xf4, 0x0, 0x55, 0x5f, 0xf6, 0x0, 0xf1, 0x75, 0x1,
		0x0, 0x9, 0x28, 0xf, 0x0, 0x94, 0xa, 0x16, 0x0, 0xc0, 0xaf, 0x15, 0x0, 0x10, 0xcb, 0xd, 0x0, 0x65, 0x80, 0x0, 0x0,
		0x5d, 0xac, 0xfd, 0x0, 0xa1, 0xe4, 0xa, 0x0, 0x83, 0x94, 0x16, 0x0, 0x39, 0x54, 0x16, 0x0, 0x1, 0x49, 0xf, 0x0,
		0x58, 0x5f, 0x6, 0x0, 0x4b, 0x48, 0x3, 0x0, 0xdb, 0x6a, 0xc, 0x0, 0x57, 0x90, 0x1a, 0x0, 0xd5, 0xcd, 0x23, 0x0,
		0xce, 0x32, 0x23, 0x0, 0xf9, 0x13, 0x14, 0x0, 0x37, 0x9d, 0xfc, 0x0, 0x3b, 0x2f, 0xef, 0x0, 0xc3, 0x85, 0xf3, 0x0,
		0xdc, 0x92, 0x1, 0x0, 0xf6, 0xe2, 0x10, 0x0, 0x3b, 0xa9, 0x1c, 0x0, 0x77, 0x4f, 0x21, 0x0, 0xb, 0x55, 0x1d, 0x0,
		0xab, 0xea, 0x10, 0x0, 0xbe, 0xa7, 0x2,
	}
	test32bitHeader = []byte{0x64, 0x61, 0x74, 0x61, 0x10, 0x92, 0x0, 0x0}
)

func BenchmarkChunk32bit(b *testing.B) {
	b.Run(
		"Parse", func(b *testing.B) {
			b.Run(
				"NewBuffer", func(b *testing.B) {
					header, err := HeaderFrom(test32bitHeader)
					if err != nil {
						b.Error(err)
						return
					}

					var chunk *Chunk32bit
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						chunk = &Chunk32bit{
							ChunkHeader: header,
							Depth:       32, // set by NewChunk()
						}
						chunk.Parse(test32bitPCM)
					}
					_ = chunk
				},
			)
			b.Run(
				"Append", func(b *testing.B) {
					header, err := HeaderFrom(test32bitHeader)
					if err != nil {
						b.Error(err)
						return
					}

					var chunk = &Chunk32bit{
						ChunkHeader: header,
						Depth:       32, // set by NewChunk()
					}
					chunk.Parse(test32bitPCM)
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						chunk.Parse(test32bitPCM)
					}
					_ = chunk
				},
			)
		},
	)
	b.Run(
		"Generate", func(b *testing.B) {
			header, err := HeaderFrom(test32bitHeader)
			if err != nil {
				b.Error(err)
				return
			}

			var (
				chunk = &Chunk32bit{
					ChunkHeader: header,
					Depth:       32, // set by NewChunk()
				}
				buf []byte
			)
			chunk.Parse(test32bitPCM)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				buf = chunk.Bytes()
			}
			_ = buf
		},
	)
}

func Test32bitHeader(t *testing.T) {
	header, err := HeaderFrom(test32bitHeader)
	if err != nil {
		t.Error(err)
		return
	}
	chunk := &Chunk32bit{
		ChunkHeader: header,
		Depth:       32, // set by NewChunk()
	}

	if output := chunk.Header(); !reflect.DeepEqual(*header, *output) {
		t.Errorf("output mismatch error: wanted %+v ; got %+v", *header, *output)
	}

	if bitDepth := chunk.BitDepth(); bitDepth != chunk.Depth {
		t.Errorf("bit depth mismatch error: wanted %v ; got %v", chunk.Depth, bitDepth)
	}
}

func Test32Bit(t *testing.T) {
	header, err := HeaderFrom(test32bitHeader)
	if err != nil {
		t.Error(err)
		return
	}

	var (
		bitDepth uint16 = 32
		input           = test32bitPCM
		chunk           = &Chunk32bit{
			ChunkHeader: header,
			Depth:       bitDepth,
		}

		f []float64
	)

	t.Run("ParseAndBytes", func(t *testing.T) {
		// clear Subchunk2Size
		chunk.Subchunk2Size = 0
		chunk.Parse(input)

		output := chunk.Bytes()
		if !bytes.Equal(input, output) {
			t.Errorf("output mismatch error: wanted %v ; got %v", input, output)
		}
	})

	t.Run("Value", func(t *testing.T) {
		if i := chunk.Value(); len(i) == 0 {
			t.Errorf("expected integer PCM buffer to be longer than zero")
		}
	})

	t.Run("Float", func(t *testing.T) {
		f = chunk.Float()
		if len(f) == 0 {
			t.Errorf("expected float PCM buffer to be longer than zero")
			return
		}
	})

	t.Run("ParseFloat", func(t *testing.T) {
		newChunk := &Chunk32bit{
			ChunkHeader: header,
		}
		newChunk.ParseFloat(f)

		if len(chunk.Data) != len(newChunk.Data) {
			t.Errorf("float data length mismatch error: wanted %d ; got %d", len(chunk.Data), len(newChunk.Data))
		}
		for i := range chunk.Data {
			if chunk.Data[i] != newChunk.Data[i] {
				t.Errorf("float data output mismatch error on index #%d: wanted %d ; got %d", i, chunk.Data[i], newChunk.Data[i])
			}
		}
	})

	t.Run("ParseSecondRun", func(t *testing.T) {
		// second run to test Parse on a dirty state
		chunk.Parse(input)
	})

	t.Run("ChunkHeader", func(t *testing.T) {
		chunkHeader := chunk.Header()
		if !reflect.DeepEqual(header, chunkHeader) {
			t.Errorf("output mismatch error: wanted %v ; got %v", header, chunkHeader)
		}
	})

	t.Run("BitDepth", func(t *testing.T) {
		depth := chunk.BitDepth()
		if depth != bitDepth {
			t.Errorf("output mismatch error: wanted %v ; got %v", bitDepth, depth)
		}
	})

	t.Run("Reset", func(t *testing.T) {
		chunk.Reset()

		if len(chunk.Data) != 0 {
			t.Errorf("output mismatch error: wanted %v ; got %v", 0, len(chunk.Data))
		}
	})

	t.Run("Generate", func(t *testing.T) {
		chunk.Generate(osc.SineWave, 2000, 16, 500*time.Millisecond)

		if len(chunk.Data) == 0 {
			t.Error("expected Data object length to be greater than zero")
		}
	})

}
