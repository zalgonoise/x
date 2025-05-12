package tracks

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed tracks.yaml
var trackBytes []byte

func TestTrackList_Read(t *testing.T) {
	for _, testcase := range []struct {
		name string
		buf  []byte
		err  error
	}{
		{
			name: "Success",
			buf:  trackBytes,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			trackList := &TrackList{}

			_, err := trackList.Read(testcase.buf)
			if err != nil {
				require.ErrorIs(t, err, testcase.err)

				return
			}

			require.NoError(t, err)
			t.Log(trackList)
		})
	}
}
