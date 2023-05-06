package data

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/zalgonoise/x/audio/osc"
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

func Test16Bit(t *testing.T) {
	header, err := HeaderFrom(test16bitHeader)
	if err != nil {
		t.Error(err)
		return
	}

	var (
		bitDepth uint16 = 16
		input           = test16bitPCM
		chunk           = &Chunk16bit{
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
		newChunk := &Chunk16bit{
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
		t.Run("Success", func(t *testing.T) {
			t.Run("SineWithNilData", func(t *testing.T) {
				chunk.Data = nil
				chunk.Generate(osc.SineWave, 2000, 44100, 100*time.Millisecond)

				if len(chunk.Data) == 0 {
					t.Error("expected Data object length to be greater than zero")
				}

				chunk.Reset()
			})
			t.Run("Square", func(t *testing.T) {
				chunk.Generate(osc.SquareWave, 2000, 44100, 100*time.Millisecond)

				if len(chunk.Data) == 0 {
					t.Error("expected Data object length to be greater than zero")
				}

				chunk.Reset()
			})
			t.Run("Triangle", func(t *testing.T) {
				chunk.Generate(osc.TriangleWave, 2000, 44100, 100*time.Millisecond)

				if len(chunk.Data) == 0 {
					t.Error("expected Data object length to be greater than zero")
				}

				chunk.Reset()
			})
			t.Run("Sawtooth", func(t *testing.T) {
				chunk.Generate(osc.SawtoothUpWave, 2000, 44100, 100*time.Millisecond)

				if len(chunk.Data) == 0 {
					t.Error("expected Data object length to be greater than zero")
				}

				chunk.Reset()
			})
		})
		t.Run("InvalidOscillatorType", func(t *testing.T) {
			chunk.Generate(osc.Type(255), 2000, 44100, 100*time.Millisecond)

			if len(chunk.Data) != 0 {
				t.Error("expected Data object length to be zero")
			}
		})
	})
}
