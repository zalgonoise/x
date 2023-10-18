package header_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	header2 "github.com/zalgonoise/x/audio/encoding/wav/data/header"
)

func TestValidate(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		input *header2.Header
		err   error
	}{
		{
			name:  "Valid/DataID",
			input: &header2.Header{Subchunk2ID: header2.Data},
		},
		{
			name:  "Valid/JunkID",
			input: &header2.Header{Subchunk2ID: header2.Junk},
		},
		{
			name:  "Invalid/UnsupportedID",
			input: &header2.Header{Subchunk2ID: [4]byte{0, 1, 2, 3}},
			err:   header2.ErrInvalidSubChunkHeader,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			err := header2.Validate(testcase.input)

			require.ErrorIs(t, err, testcase.err)
		})
	}
}
