package header_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/wav/data/header"
)

func TestValidate(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		input *header.Header
		err   error
	}{
		{
			name:  "Valid/DataID",
			input: &header.Header{Subchunk2ID: header.Data},
		},
		{
			name:  "Valid/JunkID",
			input: &header.Header{Subchunk2ID: header.Junk},
		},
		{
			name:  "Invalid/UnsupportedID",
			input: &header.Header{Subchunk2ID: [4]byte{0, 1, 2, 3}},
			err:   header.ErrInvalidSubChunkHeader,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			err := header.Validate(testcase.input)

			require.ErrorIs(t, err, testcase.err)
		})
	}
}
