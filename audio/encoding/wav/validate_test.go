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
			name: "Invalid/Subchunk1ID",
			input: func() *wav.Header {
				newHeader := *h

				newHeader.Subchunk1ID = [4]byte{'L', 'I', 'S', 'T'}

				return &newHeader
			},
			err: wav.ErrInvalidHeader,
		},
		{
			name: "Invalid/BlockAlignZero",
			input: func() *wav.Header {
				newHeader := *h

				newHeader.BlockAlign = 0

				return &newHeader
			},
			err: wav.ErrInvalidHeader,
		},
		{
			name: "Invalid/BlockAlignMismatch",
			input: func() *wav.Header {
				newHeader := *h

				newHeader.BlockAlign = 7

				return &newHeader
			},
			err: wav.ErrInvalidHeader,
		},
		{
			name: "Invalid/Float16Bit",
			input: func() *wav.Header {
				newHeader := *h

				newHeader.AudioFormat = wav.FloatFormat
				newHeader.BitsPerSample = 16
				newHeader.BlockAlign = 2 * newHeader.NumChannels

				return &newHeader
			},
			err: wav.ErrInvalidBitDepth,
		},
		{
			name: "Valid/UncommonSampleRate",
			input: func() *wav.Header {
				newHeader := *h

				newHeader.SampleRate = 3000

				return &newHeader
			},
		},
		{
			name: "Invalid/SampleRateZero",
			input: func() *wav.Header {
				newHeader := *h

				newHeader.SampleRate = 0

				return &newHeader
			},
			err: wav.ErrInvalidSampleRate,
		},
		{
			name: "Invalid/SampleRateTooHigh",
			input: func() *wav.Header {
				newHeader := *h

				newHeader.SampleRate = 2_000_000

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
			name: "Valid/FiveChannels",
			input: func() *wav.Header {
				newHeader := *h

				newHeader.NumChannels = 5
				newHeader.BlockAlign = 5 * newHeader.BitsPerSample / 8

				return &newHeader
			},
		},
		{
			name: "Invalid/NumChannelsZero",
			input: func() *wav.Header {
				newHeader := *h

				newHeader.NumChannels = 0

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

			if testcase.err == nil {
				require.NoError(t, err)

				return
			}

			require.ErrorIs(t, err, testcase.err)
		})
	}
}
