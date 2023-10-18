package wav_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/encoding/wav"
)

func TestValidate(t *testing.T) {
	h, err := wav.NewHeader(
		sampleRate, bitDepth, numChannels, audioFormat,
	)
	require.NoError(t, err)

	for _, testcase := range []struct {
		name  string
		input func() *wav.Header
		err   error
	}{
		{
			name: "Valid",
			input: func() *wav.Header {
				return h
			},
		},
		{
			name: "Invalid/NilHeader",
			input: func() *wav.Header {
				return nil
			},
			err: wav.ErrEmptyHeader,
		},
		{
			name: "Invalid/ChunkID",
			input: func() *wav.Header {
				newHeader := *h

				newHeader.ChunkID = [4]byte{0, 0, 0, 0}
				return &newHeader
			},
			err: wav.ErrInvalidHeader,
		},
		{
			name: "Invalid/Format",
			input: func() *wav.Header {
				newHeader := *h

				newHeader.Format = [4]byte{0, 0, 0, 0}
				return &newHeader
			},
			err: wav.ErrInvalidHeader,
		},
		{
			name: "Invalid/SampleRate",
			input: func() *wav.Header {
				newHeader := *h

				newHeader.SampleRate = 3000
				return &newHeader
			},
			err: wav.ErrInvalidSampleRate,
		},
		{
			name: "Invalid/BitDepth",
			input: func() *wav.Header {
				newHeader := *h

				newHeader.BitsPerSample = 4
				return &newHeader
			},
			err: wav.ErrInvalidBitDepth,
		},
		{
			name: "Invalid/NumChannels",
			input: func() *wav.Header {
				newHeader := *h

				newHeader.NumChannels = 5
				return &newHeader
			},
			err: wav.ErrInvalidNumChannels,
		},
		{
			name: "Invalid/AudioFormat",
			input: func() *wav.Header {
				newHeader := *h

				newHeader.AudioFormat = 2
				return &newHeader
			},
			err: wav.ErrInvalidAudioFormat,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			err = wav.ValidateHeader(testcase.input())

			require.ErrorIs(t, err, testcase.err)
		})
	}
}
