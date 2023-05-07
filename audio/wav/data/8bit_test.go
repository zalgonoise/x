package data

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/zalgonoise/x/audio/osc"
)

var (
	test8bitPCM = []byte{
		0x89, 0x7d, 0x75, 0x76, 0x81, 0x8f, 0x96, 0x96, 0x8e, 0x81, 0x7e, 0x8b, 0x97, 0x96, 0x8f, 0x86, 0x83, 0x8c, 0x9b,
		0xa4, 0xa3, 0x94, 0x7d, 0x6f, 0x74, 0x82, 0x91, 0x9d, 0xa1, 0x9d, 0x91, 0x83, 0x81, 0x8d, 0x96, 0x8c, 0x79, 0x73,
		0x7e, 0x92, 0x9f, 0x9c, 0x89, 0x75, 0x7a, 0x97, 0xa9, 0x9a, 0x80, 0x6b, 0x60, 0x6c, 0x89, 0xa1, 0xb6, 0xbf, 0xa1,
		0x6b, 0x4b, 0x4d, 0x5d, 0x73, 0x8f, 0xa4, 0xa9, 0xa2, 0x95, 0x8b, 0x89, 0x8c, 0x89, 0x7f, 0x78, 0x75, 0x73, 0x78,
		0x7e, 0x7e, 0x89, 0xa2, 0xac, 0x9b, 0x7e, 0x64, 0x5c, 0x6a, 0x80, 0x8f, 0x91, 0x84, 0x7c, 0x8a, 0x9c, 0x98, 0x85,
		0x6e, 0x59, 0x52, 0x59, 0x5d, 0x62, 0x79, 0x93, 0x93, 0x80, 0x74, 0x73, 0x76, 0x76, 0x6b, 0x58, 0x52, 0x67, 0x86,
		0x91, 0x88, 0x7b, 0x6f, 0x68, 0x6c, 0x71, 0x6c, 0x67, 0x6b, 0x73, 0x70, 0x60, 0x54,
	}
	test8bitHeader = []byte{0x64, 0x61, 0x74, 0x61, 0x84, 0x24, 0x0, 0x0}
)

func BenchmarkChunk8bit(b *testing.B) {
	b.Run(
		"Parse", func(b *testing.B) {
			b.Run(
				"NewBuffer", func(b *testing.B) {
					header, err := HeaderFrom(test8bitHeader)
					if err != nil {
						b.Error(err)
						return
					}

					var chunk *DataChunk
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						chunk = NewDataChunk(bitDepth8, header)
						chunk.Parse(test8bitPCM)
					}
					_ = chunk
				},
			)
			b.Run(
				"Append", func(b *testing.B) {
					header, err := HeaderFrom(test8bitHeader)
					if err != nil {
						b.Error(err)
						return
					}

					var chunk = NewDataChunk(bitDepth8, header)
					chunk.Parse(test8bitPCM)
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						chunk.Parse(test8bitPCM)
					}
					_ = chunk
				},
			)
		},
	)
	b.Run(
		"Bytes", func(b *testing.B) {
			header, err := HeaderFrom(test8bitHeader)
			if err != nil {
				b.Error(err)
				return
			}

			var (
				chunk = NewDataChunk(bitDepth8, header)
				buf   []byte
			)
			chunk.Parse(test8bitPCM)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				buf = chunk.Bytes()
			}
			_ = buf
		},
	)
}

func Test8bitHeader(t *testing.T) {
	header, err := HeaderFrom(test8bitHeader)
	if err != nil {
		t.Error(err)
		return
	}
	chunk := &DataChunk{
		ChunkHeader: header,
		Depth:       8, // set by NewChunk()
	}

	if output := chunk.Header(); !reflect.DeepEqual(*header, *output) {
		t.Errorf("output mismatch error: wanted %+v ; got %+v", *header, *output)
	}

	if bitDepth := chunk.BitDepth(); bitDepth != chunk.Depth {
		t.Errorf("bit depth mismatch error: wanted %v ; got %v", chunk.Depth, bitDepth)
	}
}

func Test8Bit(t *testing.T) {
	header, err := HeaderFrom(test8bitHeader)
	if err != nil {
		t.Error(err)
		return
	}

	var (
		bitDepth uint16 = 8
		input           = test8bitPCM
		chunk           = NewDataChunk(bitDepth, header)
		f        []float64
	)

	t.Run("ParseAndBytes", func(t *testing.T) {
		// clear Subchunk2Size
		chunk.ChunkHeader.Subchunk2Size = 0
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
		newChunk := NewDataChunk(bitDepth, header)
		newChunk.ParseFloat(f)

		if len(chunk.Data) != len(newChunk.Data) {
			t.Errorf("float data length mismatch error: wanted %d ; got %d", len(chunk.Data), len(newChunk.Data))
		}
		for i := range chunk.Data {
			if chunk.Data[i] != newChunk.Data[i] {
				t.Errorf("float data output mismatch error on index #%d: wanted %v ; got %v", i, chunk.Data[i],
					newChunk.Data[i])
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
