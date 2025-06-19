package tracks

import (
	_ "embed"
	"errors"
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
			{
				ID:           "WFHeysieGT",
				Name:         "Heysie GT",
				District:     "Waterfront",
				IsDriftTrack: true,
				CollidesWith: []string{
					"WFBudgetCup",
					"WFHeysieGTB",
					"WFNALAT",
					"WFNaimburgring",
					"WFHelenSuperGT",
				},
			},
			{
				ID:           "WFSewers",
				Name:         "Sewers",
				District:     "Waterfront",
				IsDriftTrack: false,
				CollidesWith: []string{
					"WFRailroads",
					"WFBudgetCup",
					"WFIcyGear",
					"WFBaylanBash",
					"WFNALAT",
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
		{
			name: "Fail/yaml.Unmarshal",
			buf:  []byte("tracks:\n\t- Construction:\n"),
			err:  errors.New("yaml: line 2: found character that cannot start any token"),
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			trackList := &TrackList{}

			_, err := trackList.Read(testcase.buf)
			if err != nil {
				if errors.Is(err, testcase.err) {
					return
				}

				// yaml does not provide sentinel errors to match, needs to match an error string instead
				require.Equal(t, err.Error(), testcase.err.Error())

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
		{
			name: "Fail/ErrNilList",
			err:  ErrNilList,
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
		{
			name: "Fail/ErrNilList",
			err:  ErrNilList,
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

func TestGetDriftTracks(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		list  *TrackList
		wants []string
		err   error
	}{
		{
			name: "Success",
			list: testList,
			wants: []string{
				"FinConstruction",
				"FinConstructionRev",
				"FinXmasBash",
				"WFHeysieGT",
			},
		},
		{
			name: "Fail/ErrNilList",
			err:  ErrNilList,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			tracks, err := GetDriftTracks(testcase.list)
			if err != nil {
				require.ErrorIs(t, err, testcase.err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, testcase.wants, tracks)
		})
	}
}

func TestGetTracksByDistrict(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		district  string
		driftOnly bool
		list      *TrackList
		wants     []string
		err       error
	}{
		{
			name:     "Success/All/Financial",
			district: "Financial",
			list:     testList,
			wants: []string{
				"FinConstruction",
				"FinConstructionRev",
				"FinXmasBash",
			},
		},
		{
			name:     "Success/All/Waterfront",
			district: "Waterfront",
			list:     testList,
			wants: []string{
				"WFHeysieGT",
				"WFSewers",
			},
		},
		{
			name:      "Success/DriftOnly/Financial",
			district:  "Financial",
			driftOnly: true,
			list:      testList,
			wants: []string{
				"FinConstruction",
				"FinConstructionRev",
				"FinXmasBash",
			},
		},
		{
			name:      "Success/DriftOnly/Waterfront",
			district:  "Waterfront",
			driftOnly: true,
			list:      testList,
			wants: []string{
				"WFHeysieGT",
			},
		},
		{
			name: "Fail/ErrNilList",
			err:  ErrNilList,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			tracks, err := GetTracksByDistrict(testcase.list, testcase.district, testcase.driftOnly)
			if err != nil {
				require.ErrorIs(t, err, testcase.err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, testcase.wants, tracks)
		})
	}
}

func TestGetDistricts(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		list  *TrackList
		wants []string
		err   error
	}{
		{
			name:  "Success",
			list:  testList,
			wants: []string{"Financial", "Waterfront"},
		},
		{
			name: "Fail/ErrNilList",
			err:  ErrNilList,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			districts, err := GetDistricts(testcase.list)
			if err != nil {
				require.ErrorIs(t, err, testcase.err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, testcase.wants, districts)
		})
	}
}

func TestGetNamesFromIDs(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		list  *TrackList
		ids   []string
		wants []string
		err   error
	}{
		{
			name: "Success",
			list: testList,
			ids: []string{
				"FinConstructionRev",
				"WFHeysieGT",
			},
			wants: []string{
				"Construction Rev",
				"Heysie GT",
			},
		},
		{
			name: "Fail/ErrNilList",
			err:  ErrNilList,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			names, err := GetNamesFromIDs(testcase.list, testcase.ids)
			if err != nil {
				require.ErrorIs(t, err, testcase.err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, testcase.wants, names)
		})
	}
}

func TestGetIDFromName(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		list  *TrackList
		track string
		wants string
		err   error
	}{
		{
			name:  "Success",
			list:  testList,
			track: "Construction",
			wants: "FinConstruction",
		},
		{
			name:  "Success/WithSpace",
			list:  testList,
			track: "ConstructionRev",
			wants: "FinConstructionRev",
		},
		{
			name:  "Success/LowercaseNoSpace",
			list:  testList,
			track: "xmasbash",
			wants: "FinXmasBash",
		},
		{
			name:  "Fail/NotFound",
			list:  testList,
			track: "perrytheferry",
			err:   ErrNotFound,
		},
		{
			name: "Fail/ErrNilList",
			err:  ErrNilList,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			track, err := GetIDFromName(testcase.list, testcase.track)
			if err != nil {
				require.ErrorIs(t, err, testcase.err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, testcase.wants, track)
		})
	}
}
