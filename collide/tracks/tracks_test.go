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
				ID:           "FinConstruction",
				Name:         "Construction",
				District:     "Financial",
				IsDriftTrack: true,
				CollidesWith: []string{
					"FinConstructionRev",
					"FinChurchOfPinkies",
					"FinEvasLAT",
					"FinFuckinW",
				},
			},
			{
				ID:           "FinConstructionRev",
				Name:         "Construction Rev",
				District:     "Financial",
				IsDriftTrack: true,
				CollidesWith: []string{
					"FinConstruction",
					"FinChurchOfPinkies",
					"FinEvasLAT",
					"FinFuckinW",
				},
			},
			{
				ID:           "FinXmasBash",
				Name:         "Xmas Bash",
				District:     "Financial",
				IsDriftTrack: true,
				CollidesWith: []string{
					"FinTunnelOfDeath",
					"FinLAT",
					"FinEvasLAT",
					"FinMikroTournament",
					"FinFuckinW",
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
			track: "FinConstruction",
			wants: []string{
				"FinConstructionRev",
				"FinChurchOfPinkies",
				"FinEvasLAT",
				"FinFuckinW",
			},
		},
		{
			name:  "Fail/ErrNotFound",
			list:  testList,
			track: "FinBeaconMess",
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
			track: "FinConstruction",
			wants: []string{
				"FinXmasBash",
			},
		},
		{
			name:  "Fail/ErrNotFound",
			list:  testList,
			track: "FinBeaconMess",
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
