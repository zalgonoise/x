package data

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/encoding/wav/data/filters"
	"github.com/zalgonoise/x/audio/encoding/wav/data/internal/testdata/pcm"
	"github.com/zalgonoise/x/audio/osc"
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
			header:   pcm.Test8bitHeader,
			data:     pcm.Test8bitPCM,
		},
		{
			name:     "16Bit",
			bitDepth: 16,
			header:   pcm.Test16bitHeader,
			data:     pcm.Test16bitPCM,
		},
		{
			name:     "24Bit",
			bitDepth: 24,
			header:   pcm.Test24bitHeader,
			data:     pcm.Test24bitPCM,
		},
		{
			name:     "32Bit",
			bitDepth: 32,
			header:   pcm.Test32bitHeader,
			data:     pcm.Test32bitPCM,
		},
	} {
		t.Run(class.name, func(t *testing.T) {
			h, err := From(class.header)
			require.NoError(t, err)

			for _, testcase := range []struct {
				name string
				op   func(*Chunk)
			}{
				{
					name: "ParseAndBytes",
					op: func(chunk *Chunk) {
						require.Equal(t, class.data, chunk.Bytes())
					},
				},
				{
					name: "WriteAndRead",
					op: func(chunk *Chunk) {
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
					op: func(chunk *Chunk) {
						chunk.Data = nil

						reader := bytes.NewReader(class.data)

						_, err = chunk.ReadFrom(reader)
						require.NoError(t, err)

						buf := make([]byte, len(class.data))

						_, err = chunk.Read(buf)
						require.NoError(t, err)

						require.Equal(t, class.data, buf)
					},
				},
				{
					name: "Value",
					op: func(chunk *Chunk) {
						require.Greater(t, len(chunk.Value()), 0)
					},
				},
				{
					name: "Float",
					op: func(chunk *Chunk) {
						require.Greater(t, len(chunk.Float()), 0)
					},
				},
				{
					name: "ParseFloat",
					op: func(chunk *Chunk) {
						f := chunk.Float()

						newChunk := NewPCMChunk(bitDepth16, h)
						newChunk.ParseFloat(f)
						require.Equal(t, chunk.Data, newChunk.Data)
					},
				},
				{
					name: "ParseSecondRun",
					op: func(chunk *Chunk) {
						chunk.Parse(class.data)
					},
				},
				{
					name: "Header",
					op: func(chunk *Chunk) {
						require.Equal(t, h, chunk.Header())
					},
				},
				{
					name: "BitDepth",
					op: func(chunk *Chunk) {
						require.Equal(t, class.bitDepth, chunk.BitDepth())
					},
				},
				{
					name: "Reset",
					op: func(chunk *Chunk) {
						chunk.Reset()
						require.Len(t, chunk.Data, 0)
					},
				},
				{
					name: "Generate/Success/SineWithNilData",
					op: func(chunk *Chunk) {
						chunk.Data = nil
						chunk.Generate(osc.SineWave, 2000, 44100, 100*time.Millisecond)

						require.Greater(t, len(chunk.Data), 0)
					},
				},
				{
					name: "Generate/Success/Square",
					op: func(chunk *Chunk) {
						chunk.Data = nil
						chunk.Generate(osc.SquareWave, 2000, 44100, 100*time.Millisecond)

						require.Greater(t, len(chunk.Data), 0)
					},
				},
				{
					name: "Generate/Success/Triangle",
					op: func(chunk *Chunk) {
						chunk.Data = nil
						chunk.Generate(osc.TriangleWave, 2000, 44100, 100*time.Millisecond)

						require.Greater(t, len(chunk.Data), 0)
					},
				},
				{
					name: "Generate/Success/SawtoothUp",
					op: func(chunk *Chunk) {
						chunk.Data = nil
						chunk.Generate(osc.SawtoothUpWave, 2000, 44100, 100*time.Millisecond)

						require.Greater(t, len(chunk.Data), 0)
					},
				},
				{
					name: "Generate/Success/SawtoothDown",
					op: func(chunk *Chunk) {
						chunk.Data = nil
						chunk.Generate(osc.SawtoothDownWave, 2000, 44100, 100*time.Millisecond)

						require.Greater(t, len(chunk.Data), 0)
					},
				},
				{
					name: "Generate/Fail",
					op: func(chunk *Chunk) {
						chunk.Data = nil
						chunk.Generate(osc.Type(255), 2000, 44100, 100*time.Millisecond)

						require.Len(t, chunk.Data, 0)
					},
				},
				{
					name: "Apply",
					op: func(chunk *Chunk) {
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
					chunk := NewPCMChunk(class.bitDepth, h)

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
			header:   pcm.Test8bitHeader,
			data:     pcm.Test8bitPCM,
			size:     64,
		},
		{
			name:     "16Bit",
			bitDepth: 16,
			header:   pcm.Test16bitHeader,
			data:     pcm.Test16bitPCM,
			size:     64,
		},
		{
			name:     "24Bit",
			bitDepth: 24,
			header:   pcm.Test24bitHeader,
			data:     pcm.Test24bitPCM,
			size:     96,
		},
		{
			name:     "32Bit",
			bitDepth: 32,
			header:   pcm.Test32bitHeader,
			data:     pcm.Test32bitPCM,
			size:     128,
		},
	} {
		t.Run(class.name, func(t *testing.T) {
			h, err := From(class.header)
			require.NoError(t, err)

			for _, testcase := range []struct {
				name string
				op   func(*Ring)
			}{
				{
					name: "ParseAndBytes",
					op: func(chunk *Ring) {
						require.Equal(t, class.data[len(class.data)-class.size:], chunk.Bytes())
					},
				},
				{
					name: "WriteAndRead",
					op: func(chunk *Ring) {
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
					op: func(chunk *Ring) {
						chunk.Data.Reset()

						reader := bytes.NewReader(class.data)

						_, err = chunk.ReadFrom(reader)
						require.NoError(t, err)

						buf := make([]byte, class.size)

						_, err = chunk.Read(buf)
						require.NoError(t, err)

						require.Equal(t, class.data[len(class.data)-class.size:], buf)
					},
				},
				{
					name: "Value",
					op: func(chunk *Ring) {
						require.Len(t, chunk.Value(), class.size/(int(class.bitDepth)/byteSize))
					},
				},
				{
					name: "Float",
					op: func(chunk *Ring) {
						require.Len(t, chunk.Float(), class.size/(int(class.bitDepth)/byteSize))
					},
				},
				{
					name: "ParseFloat",
					op: func(chunk *Ring) {
						f := chunk.Float()

						newChunk := NewPCMRing(bitDepth16, h, class.size*2, nil)
						newChunk.ParseFloat(f)
						require.Equal(t, chunk.Data.Value(), newChunk.Data.Value())
					},
				},
				{
					name: "ParseSecondRun",
					op: func(chunk *Ring) {
						chunk.Parse(class.data)
					},
				},
				{
					name: "Header",
					op: func(chunk *Ring) {
						require.Equal(t, h, chunk.Header())
					},
				},
				{
					name: "BitDepth",
					op: func(chunk *Ring) {
						require.Equal(t, class.bitDepth, chunk.BitDepth())
					},
				},
				{
					name: "Reset",
					op: func(chunk *Ring) {
						chunk.Reset()
					},
				},
				{
					name: "Generate/Success/SineWithNilData",
					op: func(chunk *Ring) {
						chunk.Generate(osc.SineWave, 2000, 44100, 100*time.Millisecond)

						require.Len(t, chunk.Data.Value(), class.size/(int(class.bitDepth)/byteSize))
					},
				},
				{
					name: "Generate/Success/Square",
					op: func(chunk *Ring) {
						chunk.Generate(osc.SquareWave, 2000, 44100, 100*time.Millisecond)

						require.Len(t, chunk.Data.Value(), class.size/(int(class.bitDepth)/byteSize))
					},
				},
				{
					name: "Generate/Success/Triangle",
					op: func(chunk *Ring) {
						chunk.Generate(osc.TriangleWave, 2000, 44100, 100*time.Millisecond)

						require.Len(t, chunk.Data.Value(), class.size/(int(class.bitDepth)/byteSize))
					},
				},
				{
					name: "Generate/Success/SawtoothUp",
					op: func(chunk *Ring) {
						chunk.Generate(osc.SawtoothUpWave, 2000, 44100, 100*time.Millisecond)

						require.Len(t, chunk.Data.Value(), class.size/(int(class.bitDepth)/byteSize))
					},
				},
				{
					name: "Generate/Success/SawtoothDown",
					op: func(chunk *Ring) {
						chunk.Generate(osc.SawtoothDownWave, 2000, 44100, 100*time.Millisecond)

						require.Len(t, chunk.Data.Value(), class.size/(int(class.bitDepth)/byteSize))
					},
				},
				{
					name: "Generate/Fail",
					op: func(chunk *Ring) {
						chunk.Generate(osc.Type(255), 2000, 44100, 100*time.Millisecond)
					},
				},
				{
					name: "Apply",
					op: func(chunk *Ring) {
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
					chunk := NewPCMRing(class.bitDepth, h, class.size, nil)

					chunk.Parse(class.data)
					testcase.op(chunk)
				})
			}
		})
	}
}
