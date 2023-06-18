package data

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/osc"
	"github.com/zalgonoise/x/audio/wav/data/header"
)

func TestDataChunk(t *testing.T) {
	for _, class := range []struct {
		name     string
		bitDepth uint16
		header   []byte
		data     []byte
	}{
		{
			name:     "8Bit",
			bitDepth: 8,
			header:   test8bitHeader,
			data:     test8bitPCM,
		},
		{
			name:     "16Bit",
			bitDepth: 16,
			header:   test16bitHeader,
			data:     test16bitPCM,
		},
		{
			name:     "24Bit",
			bitDepth: 24,
			header:   test24bitHeader,
			data:     test24bitPCM,
		},
		{
			name:     "32Bit",
			bitDepth: 32,
			header:   test32bitHeader,
			data:     test32bitPCM,
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
					name: "WriteAndRead",
					op: func(chunk *DataChunk) {
						chunk.Data = nil

						_, err = chunk.Write(class.data)
						require.NoError(t, err)

						buf := make([]byte, len(class.data))

						_, err = chunk.Read(buf)
						require.NoError(t, err)

						require.Equal(t, class.data, buf)
					},
				},
				{
					name: "ReadFrom",
					op: func(chunk *DataChunk) {
						chunk.Data = nil

						reader := bytes.NewReader(class.data)

						_, err = chunk.ReadFrom(reader)
						require.NoError(t, err)

						buf := make([]byte, len(class.data))

						_, err = chunk.Read(buf)
						require.NoError(t, err)

						require.Equal(t, class.data, buf)
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
