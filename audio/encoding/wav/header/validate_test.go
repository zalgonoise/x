package header_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	header2 "github.com/zalgonoise/x/audio/encoding/wav/header"
)

func TestValidate(t *testing.T) {
	h, err := header2.New(
		sampleRate, bitDepth, numChannels, audioFormat,
	)
	require.NoError(t, err)

	for _, testcase := range []struct {
		name  string
		input func() *header2.Header
		err   error
	}{
		{
			name: "Valid",
			input: func() *header2.Header {
				return h
			},
		},
		{
			name: "Invalid/NilHeader",
			input: func() *header2.Header {
				return nil
			},
			err: header2.ErrEmptyHeader,
		},
		{
			name: "Invalid/ChunkID",
			input: func() *header2.Header {
				newHeader := *h

				newHeader.ChunkID = [4]byte{0, 0, 0, 0}
				return &newHeader
			},
			err: header2.ErrInvalidHeader,
		},
		{
			name: "Invalid/Format",
			input: func() *header2.Header {
				newHeader := *h

				newHeader.Format = [4]byte{0, 0, 0, 0}
				return &newHeader
			},
			err: header2.ErrInvalidHeader,
		},
		{
			name: "Invalid/SampleRate",
			input: func() *header2.Header {
				newHeader := *h

				newHeader.SampleRate = 3000
				return &newHeader
			},
			err: header2.ErrInvalidSampleRate,
		},
		{
			name: "Invalid/BitDepth",
			input: func() *header2.Header {
				newHeader := *h

				newHeader.BitsPerSample = 4
				return &newHeader
			},
			err: header2.ErrInvalidBitDepth,
		},
		{
			name: "Invalid/NumChannels",
			input: func() *header2.Header {
				newHeader := *h

				newHeader.NumChannels = 5
				return &newHeader
			},
			err: header2.ErrInvalidNumChannels,
		},
		{
			name: "Invalid/AudioFormat",
			input: func() *header2.Header {
				newHeader := *h

				newHeader.AudioFormat = 2
				return &newHeader
			},
			err: header2.ErrInvalidAudioFormat,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			err = header2.Validate(testcase.input())

			require.ErrorIs(t, err, testcase.err)
		})
	}
}
