package data

import (
	"bytes"
	"github.com/zalgonoise/x/audio/wav/data/filters"
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
				}, {
					name: "SetBitDepth/RoundTrip",
					op: func(chunk *DataChunk) {
						origDepth := chunk.Depth

						newChunk, err := chunk.SetBitDepth(16)
						require.NoError(t, err)

						rtChunk, err := newChunk.SetBitDepth(origDepth)
						require.NoError(t, err)

						require.Equal(t, chunk.Data, rtChunk.Data)
					},
				}, {
					name: "Apply",
					op: func(chunk *DataChunk) {
						orig := make([]float64, len(chunk.Data))
						copy(orig, chunk.Data)

						chunk.Apply(
							filters.PhaseFlip(),
							filters.PhaseFlip(),
						)

						require.Equal(t, orig, chunk.Data)
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

func TestDataRing(t *testing.T) {
	for _, class := range []struct {
		name     string
		bitDepth uint16
		header   []byte
		data     []byte
		size     int
	}{
		{
			name:     "8Bit",
			bitDepth: 8,
			header:   test8bitHeader,
			data:     test8bitPCM,
			size:     64,
		},
		{
			name:     "16Bit",
			bitDepth: 16,
			header:   test16bitHeader,
			data:     test16bitPCM,
			size:     64,
		},
		{
			name:     "24Bit",
			bitDepth: 24,
			header:   test24bitHeader,
			data:     test24bitPCM,
			size:     96,
		},
		{
			name:     "32Bit",
			bitDepth: 32,
			header:   test32bitHeader,
			data:     test32bitPCM,
			size:     128,
		},
	} {
		t.Run(class.name, func(t *testing.T) {
			h, err := header.From(class.header)
			require.NoError(t, err)

			for _, testcase := range []struct {
				name string
				op   func(*DataRing)
			}{
				{
					name: "ParseAndBytes",
					op: func(chunk *DataRing) {
						require.Equal(t, class.data[len(class.data)-class.size:], chunk.Bytes())
					},
				}, {
					name: "WriteAndRead",
					op: func(chunk *DataRing) {
						chunk.Data.Reset()

						_, err = chunk.Write(class.data)
						require.NoError(t, err)

						buf := make([]byte, class.size)

						_, err = chunk.Read(buf)
						require.NoError(t, err)

						require.Equal(t, class.data[len(class.data)-class.size:], buf)
					},
				},
				{
					name: "ReadFrom",
					op: func(chunk *DataRing) {
						chunk.Data.Reset()

						reader := bytes.NewReader(class.data)

						_, err = chunk.ReadFrom(reader)
						require.NoError(t, err)

						buf := make([]byte, class.size)

						_, err = chunk.Read(buf)
						require.NoError(t, err)

						require.Equal(t, class.data[len(class.data)-class.size:], buf)
					},
				}, {
					name: "Value",
					op: func(chunk *DataRing) {
						require.Len(t, chunk.Value(), class.size/(int(class.bitDepth)/byteSize))
					},
				}, {
					name: "Float",
					op: func(chunk *DataRing) {
						require.Len(t, chunk.Float(), class.size/(int(class.bitDepth)/byteSize))
					},
				}, {
					name: "ParseFloat",
					op: func(chunk *DataRing) {
						f := chunk.Float()

						newChunk := NewPCMDataRing(class.size*2, bitDepth16, h)
						newChunk.ParseFloat(f)
						require.Equal(t, chunk.Data.Value(), newChunk.Data.Value())
					},
				}, {
					name: "ParseSecondRun",
					op: func(chunk *DataRing) {
						chunk.Parse(class.data)
					},
				}, {
					name: "Header",
					op: func(chunk *DataRing) {
						require.Equal(t, h, chunk.Header())
					},
				}, {
					name: "BitDepth",
					op: func(chunk *DataRing) {
						require.Equal(t, class.bitDepth, chunk.BitDepth())
					},
				}, {
					name: "Reset",
					op: func(chunk *DataRing) {
						chunk.Reset()
					},
				}, {
					name: "Generate/Success/SineWithNilData",
					op: func(chunk *DataRing) {
						chunk.Generate(osc.SineWave, 2000, 44100, 100*time.Millisecond)

						require.Len(t, chunk.Data.Value(), class.size/(int(class.bitDepth)/byteSize))
					},
				}, {
					name: "Generate/Success/Square",
					op: func(chunk *DataRing) {
						chunk.Generate(osc.SquareWave, 2000, 44100, 100*time.Millisecond)

						require.Len(t, chunk.Data.Value(), class.size/(int(class.bitDepth)/byteSize))
					},
				}, {
					name: "Generate/Success/Triangle",
					op: func(chunk *DataRing) {
						chunk.Generate(osc.TriangleWave, 2000, 44100, 100*time.Millisecond)

						require.Len(t, chunk.Data.Value(), class.size/(int(class.bitDepth)/byteSize))
					},
				}, {
					name: "Generate/Success/SawtoothUp",
					op: func(chunk *DataRing) {
						chunk.Generate(osc.SawtoothUpWave, 2000, 44100, 100*time.Millisecond)

						require.Len(t, chunk.Data.Value(), class.size/(int(class.bitDepth)/byteSize))
					},
				}, {
					name: "Generate/Success/SawtoothDown",
					op: func(chunk *DataRing) {
						chunk.Generate(osc.SawtoothDownWave, 2000, 44100, 100*time.Millisecond)

						require.Len(t, chunk.Data.Value(), class.size/(int(class.bitDepth)/byteSize))
					},
				}, {
					name: "Generate/Fail",
					op: func(chunk *DataRing) {
						chunk.Generate(osc.Type(255), 2000, 44100, 100*time.Millisecond)
					},
				}, {
					name: "SetBitDepth/RoundTrip",
					op: func(chunk *DataRing) {
						//origDepth := chunk.Depth
						//
						//newChunk, err := chunk.SetBitDepth(16)
						//require.NoError(t, err)
						//
						//rtChunk, err := newChunk.SetBitDepth(origDepth)
						//require.NoError(t, err)
						//
						//require.Equal(t, chunk.Data, rtChunk.Data)
					},
				}, {
					name: "Apply",
					op: func(chunk *DataRing) {
						orig := make([]float64, chunk.Data.Cap())
						copy(orig, chunk.Data.Value())

						chunk.Apply(
							filters.PhaseFlip(),
							filters.PhaseFlip(),
						)

						require.Equal(t, orig, chunk.Data.Value())
					},
				},
			} {
				t.Run(testcase.name, func(t *testing.T) {
					chunk := NewPCMDataRing(class.size, class.bitDepth, h)
					chunk.Parse(class.data)
					testcase.op(chunk)
				})
			}
		})
	}
}
