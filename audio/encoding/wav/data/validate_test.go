package data_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zalgonoise/x/audio/encoding/wav/data"
)

func TestValidate(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		input *data.Header
		err   error
	}{
		{
			name:  "Valid/DataID",
			input: &data.Header{Subchunk2ID: data.Data},
		},
		{
			name:  "Valid/JunkID",
			input: &data.Header{Subchunk2ID: data.Junk},
		},
		{
			name:  "Invalid/UnsupportedID",
			input: &data.Header{Subchunk2ID: [4]byte{0, 1, 2, 3}},
			err:   data.ErrInvalidSubChunkHeader,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			err := data.Validate(testcase.input)

			require.ErrorIs(t, err, testcase.err)
		})
	}
}
