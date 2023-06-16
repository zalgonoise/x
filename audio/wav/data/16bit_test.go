package data

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/osc"
	"github.com/zalgonoise/x/audio/wav/data/header"
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
					h, err := header.From(test16bitHeader)
					if err != nil {
						b.Error(err)
						return
					}

					var chunk *DataChunk
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						chunk = NewPCMDataChunk(bitDepth16, h)
						chunk.Parse(test16bitPCM)
					}
					_ = chunk
				},
			)
			b.Run(
				"Append", func(b *testing.B) {
					h, err := header.From(test16bitHeader)
					if err != nil {
						b.Error(err)
						return
					}

					var chunk = NewPCMDataChunk(bitDepth16, h)
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
			h, err := header.From(test16bitHeader)
			if err != nil {
				b.Error(err)
				return
			}

			var (
				chunk = NewPCMDataChunk(bitDepth16, h)
				buf   []byte
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
	h, err := header.From(test16bitHeader)
	if err != nil {
		t.Error(err)
		return
	}
	chunk := NewPCMDataChunk(bitDepth16, h)

	if output := chunk.Header(); !reflect.DeepEqual(*h, *output) {
		t.Errorf("output mismatch error: wanted %+v ; got %+v", *h, *output)
	}

	if bitDepth := chunk.BitDepth(); bitDepth != chunk.Depth {
		t.Errorf("bit depth mismatch error: wanted %v ; got %v", chunk.Depth, bitDepth)
	}
}

func TestData(t *testing.T) {
	for _, class := range []struct {
		name     string
		bitDepth uint16
		header   []byte
		data     []byte
	}{
		{
			name:     "16Bit",
			bitDepth: 16,
			header:   test16bitHeader,
			data:     test16bitPCM,
		},
	} {
		t.Run(class.name, func(t *testing.T) {
			h, err := header.From(class.header)
			require.NoError(t, err)

			for _, testcase := range []struct {
				name string
				op   func(*DataChunk)
			}{
				{
					name: "ParseAndBytes",
					op: func(chunk *DataChunk) {
						require.Equal(t, class.data, chunk.Bytes())
					},
				}, {
					name: "Value",
					op: func(chunk *DataChunk) {
						require.Greater(t, len(chunk.Value()), 0)
					},
				}, {
					name: "Float",
					op: func(chunk *DataChunk) {
						require.Greater(t, len(chunk.Float()), 0)
					},
				}, {
					name: "ParseFloat",
					op: func(chunk *DataChunk) {
						f := chunk.Float()

						newChunk := NewPCMDataChunk(bitDepth16, h)
						newChunk.ParseFloat(f)
						require.Equal(t, chunk.Data, newChunk.Data)
					},
				}, {
					name: "ParseSecondRun",
					op: func(chunk *DataChunk) {
						chunk.Parse(class.data)
					},
				}, {
					name: "Header",
					op: func(chunk *DataChunk) {
						require.Equal(t, h, chunk.Header())
					},
				}, {
					name: "BitDepth",
					op: func(chunk *DataChunk) {
						require.Equal(t, class.bitDepth, chunk.BitDepth())
					},
				}, {
					name: "Reset",
					op: func(chunk *DataChunk) {
						chunk.Reset()
						require.Len(t, chunk.Data, 0)
					},
				}, {
					name: "Generate/Success/SineWithNilData",
					op: func(chunk *DataChunk) {
						chunk.Data = nil
						chunk.Generate(osc.SineWave, 2000, 44100, 100*time.Millisecond)

						require.Greater(t, len(chunk.Data), 0)
					},
				}, {
					name: "Generate/Success/Square",
					op: func(chunk *DataChunk) {
						chunk.Data = nil
						chunk.Generate(osc.SquareWave, 2000, 44100, 100*time.Millisecond)

						require.Greater(t, len(chunk.Data), 0)
					},
				}, {
					name: "Generate/Success/Triangle",
					op: func(chunk *DataChunk) {
						chunk.Data = nil
						chunk.Generate(osc.TriangleWave, 2000, 44100, 100*time.Millisecond)

						require.Greater(t, len(chunk.Data), 0)
					},
				}, {
					name: "Generate/Success/SawtoothUp",
					op: func(chunk *DataChunk) {
						chunk.Data = nil
						chunk.Generate(osc.SawtoothUpWave, 2000, 44100, 100*time.Millisecond)

						require.Greater(t, len(chunk.Data), 0)
					},
				}, {
					name: "Generate/Success/SawtoothDown",
					op: func(chunk *DataChunk) {
						chunk.Data = nil
						chunk.Generate(osc.SawtoothDownWave, 2000, 44100, 100*time.Millisecond)

						require.Greater(t, len(chunk.Data), 0)
					},
				}, {
					name: "Generate/Fail",
					op: func(chunk *DataChunk) {
						chunk.Data = nil
						chunk.Generate(osc.Type(255), 2000, 44100, 100*time.Millisecond)

						require.Len(t, chunk.Data, 0)
					},
				},
			} {
				t.Run(testcase.name, func(t *testing.T) {
					chunk := NewPCMDataChunk(class.bitDepth, h)
					chunk.Parse(class.data)
					testcase.op(chunk)
				})
			}
		})
	}
}
