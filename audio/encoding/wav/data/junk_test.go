package data

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/encoding/wav/data/filters"
	"github.com/zalgonoise/x/audio/osc"
)

func TestJunkChunk(t *testing.T) {
	var (
		err   error
		input = []byte("some junk data")
		h     = &Header{
			Subchunk2ID:   JunkID,
			Subchunk2Size: 14,
		}
		bitDepth = 0
	)

	for _, testcase := range []struct {
		name string
		op   func(chunk *Junk)
	}{
		{
			name: "ParseAndBytes",
			op: func(chunk *Junk) {
				require.Equal(t, input, chunk.Bytes())
			},
		},
		{
			name: "WriteAndRead",
			op: func(chunk *Junk) {
				chunk.Data = nil

				_, err = chunk.Write(input)
				require.NoError(t, err)

				buf := make([]byte, len(input))

				_, err = chunk.Read(buf)
				require.NoError(t, err)

				require.Equal(t, input, buf)
			},
		},
		{
			name: "ReadFrom",
			op: func(chunk *Junk) {
				chunk.Data = nil

				reader := bytes.NewReader(input)

				_, err = chunk.ReadFrom(reader)
				require.NoError(t, err)

				buf := make([]byte, len(input))

				_, err = chunk.Read(buf)
				require.NoError(t, err)

				require.Equal(t, input, buf)
			},
		},
		{
			name: "Value",
			op: func(chunk *Junk) {
				require.Greater(t, len(chunk.Value()), 0)
			},
		},
		{
			name: "Float",
			op: func(chunk *Junk) {
				require.Equal(t, len(chunk.Float()), 0)
			},
		},
		{
			name: "ParseFloat",
			op: func(chunk *Junk) {
				chunk.ParseFloat([]float64{1.5})
			},
		},
		{
			name: "ParseSecondRun",
			op: func(chunk *Junk) {
				chunk.Parse(input)
			},
		},
		{
			name: "Header",
			op: func(chunk *Junk) {
				require.Equal(t, h, chunk.Header())
			},
		},
		{
			name: "BitDepth",
			op: func(chunk *Junk) {
				require.Equal(t, uint16(bitDepth), chunk.BitDepth())
			},
		},
		{
			name: "Reset",
			op: func(chunk *Junk) {
				chunk.Reset()
				require.Len(t, chunk.Data, 0)
			},
		},
		{
			name: "Generate/Success/SineWithNilData",
			op: func(chunk *Junk) {
				chunk.Data = nil
				chunk.Generate(osc.SineWave, 2000, 44100, 100*time.Millisecond)

				require.Len(t, chunk.Data, 0)
			},
		},
		{
			name: "Generate/Success/Square",
			op: func(chunk *Junk) {
				chunk.Data = nil
				chunk.Generate(osc.SquareWave, 2000, 44100, 100*time.Millisecond)

				require.Len(t, chunk.Data, 0)
			},
		},
		{
			name: "Generate/Success/Triangle",
			op: func(chunk *Junk) {
				chunk.Data = nil
				chunk.Generate(osc.TriangleWave, 2000, 44100, 100*time.Millisecond)

				require.Len(t, chunk.Data, 0)
			},
		},
		{
			name: "Generate/Success/SawtoothUp",
			op: func(chunk *Junk) {
				chunk.Data = nil
				chunk.Generate(osc.SawtoothUpWave, 2000, 44100, 100*time.Millisecond)

				require.Len(t, chunk.Data, 0)
			},
		},
		{
			name: "Generate/Success/SawtoothDown",
			op: func(chunk *Junk) {
				chunk.Data = nil
				chunk.Generate(osc.SawtoothDownWave, 2000, 44100, 100*time.Millisecond)

				require.Len(t, chunk.Data, 0)
			},
		},
		{
			name: "Generate/Fail",
			op: func(chunk *Junk) {
				chunk.Data = nil
				chunk.Generate(osc.Type(255), 2000, 44100, 100*time.Millisecond)

				require.Len(t, chunk.Data, 0)
			},
		},
		{
			name: "Apply",
			op: func(chunk *Junk) {
				orig := make([]byte, len(chunk.Data))
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
			chunk := NewJunk(h)
			chunk.Parse(input)
			testcase.op(chunk)
		})
	}
}
