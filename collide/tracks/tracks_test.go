package tracks

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	//go:embed tracks.yaml
	trackBytes []byte

	testList = &TrackList{
		Tracks: []Track{
			{
				ID:           0,
				Name:         "Construction",
				District:     "Financial",
				IsDriftTrack: true,
				CollidesWith: []string{
					"Construction Rev",
					"Church Of Pinkies",
					"Eva's LAT",
					"Fuckin' W",
				},
			},
			{
				ID:           1,
				Name:         "Construction Rev",
				District:     "Financial",
				IsDriftTrack: true,
				CollidesWith: []string{
					"Construction",
					"Church Of Pinkies",
					"Eva's LAT",
					"Fuckin' W",
				},
			},
			{
				ID:           2,
				Name:         "Xmas Bash",
				District:     "Financial",
				IsDriftTrack: true,
				CollidesWith: []string{
					"Tunnel of Death",
					"LAT",
					"Eva's LAT",
					"Mikro Tournament",
					"Fuckin' W",
				},
			},
		},
	}
)

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

func TestGetCollisions(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		list  *TrackList
		track string
		wants []string
		err   error
	}{
		{
			name:  "Success",
			list:  testList,
			track: "Construction",
			wants: []string{
				"Construction Rev",
				"Church Of Pinkies",
				"Eva's LAT",
				"Fuckin' W",
			},
		},
		{
			name:  "Fail/ErrNotFound",
			list:  testList,
			track: "Beacon Mess",
			err:   ErrNotFound,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			tracks, err := GetCollisions(testcase.list, testcase.track)
			if err != nil {
				require.ErrorIs(t, err, testcase.err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, testcase.wants, tracks)
		})
	}
}

func TestGetOpenTracks(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		list  *TrackList
		track string
		wants []string
		err   error
	}{
		{
			name:  "Success",
			list:  testList,
			track: "Construction",
			wants: []string{
				"Xmas Bash",
			},
		},
		{
			name:  "Fail/ErrNotFound",
			list:  testList,
			track: "Beacon Mess",
			err:   ErrNotFound,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			tracks, err := GetOpenTracks(testcase.list, testcase.track)
			if err != nil {
				require.ErrorIs(t, err, testcase.err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, testcase.wants, tracks)
		})
	}
}
