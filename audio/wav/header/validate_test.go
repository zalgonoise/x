package header_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/wav/header"
)

func TestValidate(t *testing.T) {
	h, err := header.New(
		sampleRate, bitDepth, numChannels, audioFormat,
	)
	require.NoError(t, err)

	for _, testcase := range []struct {
		name  string
		input func() *header.Header
		err   error
	}{
		{
			name: "Valid",
			input: func() *header.Header {
				return h
			},
		},
		{
			name: "Invalid/NilHeader",
			input: func() *header.Header {
				return nil
			},
			err: header.ErrEmptyHeader,
		},
		{
			name: "Invalid/ChunkID",
			input: func() *header.Header {
				newHeader := *h

				newHeader.ChunkID = [4]byte{0, 0, 0, 0}
				return &newHeader
			},
			err: header.ErrInvalidHeader,
		},
		{
			name: "Invalid/Format",
			input: func() *header.Header {
				newHeader := *h

				newHeader.Format = [4]byte{0, 0, 0, 0}
				return &newHeader
			},
			err: header.ErrInvalidHeader,
		},
		{
			name: "Invalid/SampleRate",
			input: func() *header.Header {
				newHeader := *h

				newHeader.SampleRate = 3000
				return &newHeader
			},
			err: header.ErrInvalidSampleRate,
		},
		{
			name: "Invalid/BitDepth",
			input: func() *header.Header {
				newHeader := *h

				newHeader.BitsPerSample = 4
				return &newHeader
			},
			err: header.ErrInvalidBitDepth,
		},
		{
			name: "Invalid/NumChannels",
			input: func() *header.Header {
				newHeader := *h

				newHeader.NumChannels = 5
				return &newHeader
			},
			err: header.ErrInvalidNumChannels,
		},
		{
			name: "Invalid/AudioFormat",
			input: func() *header.Header {
				newHeader := *h

				newHeader.AudioFormat = 2
				return &newHeader
			},
			err: header.ErrInvalidAudioFormat,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			// implied Validate call on return
			err := header.Validate(testcase.input())

			require.ErrorIs(t, err, testcase.err)
		})
	}
}
