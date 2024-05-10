package mapping_test

import (
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zalgonoise/x/mapping"
)

type user struct {
	id   int
	name string
}

type data struct {
	len int
}

type blob struct {
	user user
	data data
}

type dataSet struct {
	interval mapping.Interval
	blob     blob
}

func TestTimeframeReplacer(t *testing.T) {
	interval1 := mapping.Interval{
		From: time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 18, 0, 0, 0, time.UTC),
	}

	interval2 := mapping.Interval{
		From: time.Date(2024, 1, 10, 18, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 21, 0, 0, 0, time.UTC),
	}

	interval3 := mapping.Interval{
		From: time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 21, 0, 0, 0, time.UTC),
	}

	interval4 := mapping.Interval{
		From: time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
	}

	i4split := mapping.Interval{
		From: time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 18, 0, 0, 0, time.UTC),
	}

	interval5 := mapping.Interval{
		From: time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 17, 0, 0, 0, time.UTC),
	}

	i5split1 := mapping.Interval{
		From: time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
	}

	i5split2 := mapping.Interval{
		From: time.Date(2024, 1, 10, 17, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 18, 0, 0, 0, time.UTC),
	}

	interval6 := mapping.Interval{
		From: time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 23, 0, 0, 0, time.UTC),
	}

	i6split := mapping.Interval{
		From: time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
	}

	interval7 := mapping.Interval{
		From: time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 18, 0, 0, 0, time.UTC),
	}

	i7split := mapping.Interval{
		From: time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
	}

	interval8 := mapping.Interval{
		From: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 20, 0, 0, 0, time.UTC),
	}

	interval9 := mapping.Interval{
		From: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
	}

	i9split := mapping.Interval{
		From: time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 18, 0, 0, 0, time.UTC),
	}

	interval10 := mapping.Interval{
		From: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 18, 0, 0, 0, time.UTC),
	}

	blob1 := blob{
		user: user{0, "user"},
		data: data{63},
	}

	blob2 := blob{
		user: user{0, "user_x"},
		data: data{64},
	}

	for _, testcase := range []struct {
		name  string
		sets  []dataSet
		wants []dataSet
	}{
		{
			name:  "OneBlob",
			sets:  []dataSet{{interval: interval1, blob: blob1}},
			wants: []dataSet{{interval: interval1, blob: blob1}},
		},
		{
			name: "TwoBlobs/Separate/NextIsAfter",
			sets: []dataSet{
				{interval: interval1, blob: blob1},
				{interval: interval2, blob: blob2},
			},
			wants: []dataSet{
				{interval: interval1, blob: blob1},
				{interval: interval2, blob: blob2},
			},
		},
		{
			name: "TwoBlobs/Separate/NextIsBefore",
			sets: []dataSet{
				{interval: interval2, blob: blob2},
				{interval: interval1, blob: blob1},
			},
			wants: []dataSet{
				{interval: interval1, blob: blob1},
				{interval: interval2, blob: blob2},
			},
		},
		{
			name: "TwoBlobs/MatchingStart/NextOverlapsCurrent",
			sets: []dataSet{
				{interval: interval1, blob: blob1},
				{interval: interval3, blob: blob2},
			},
			wants: []dataSet{
				{interval: interval3, blob: blob2},
			},
		},
		{
			name: "TwoBlobs/MatchingStart/NextOverlapsCurrent",
			sets: []dataSet{
				{interval: interval1, blob: blob1},
				{interval: interval3, blob: blob2},
			},
			wants: []dataSet{
				{interval: interval3, blob: blob2},
			},
		},
		{
			name: "TwoBlobs/MatchingStart/NextWithinCurrent",
			sets: []dataSet{
				{interval: interval1, blob: blob1},
				{interval: interval4, blob: blob2},
			},
			wants: []dataSet{
				{interval: interval4, blob: blob2},
				{interval: i4split, blob: blob1},
			},
		},
		{
			name: "TwoBlobs/MatchingStart/NextIsSameRange",
			sets: []dataSet{
				{interval: interval1, blob: blob1},
				{interval: interval1, blob: blob2},
			},
			wants: []dataSet{
				{interval: interval1, blob: blob2},
			},
		},
		{
			name: "TwoBlobs/OverlappingMiddle/NextWithinCurrent",
			sets: []dataSet{
				{interval: interval1, blob: blob1},
				{interval: interval5, blob: blob2},
			},
			wants: []dataSet{
				{interval: i5split1, blob: blob1},
				{interval: interval5, blob: blob2},
				{interval: i5split2, blob: blob1},
			},
		},
		{
			name: "TwoBlobs/OverlappingEnd/NextGoesBeyondCurrent",
			sets: []dataSet{
				{interval: interval1, blob: blob1},
				{interval: interval6, blob: blob2},
			},
			wants: []dataSet{
				{interval: i6split, blob: blob1},
				{interval: interval6, blob: blob2},
			},
		},
		{
			name: "TwoBlobs/OverlappingEnd/NextMatchesEnds",
			sets: []dataSet{
				{interval: interval1, blob: blob1},
				{interval: interval7, blob: blob2},
			},
			wants: []dataSet{
				{interval: i7split, blob: blob1},
				{interval: interval7, blob: blob2},
			},
		},
		{
			name: "TwoBlobs/OverlappingStart/NextCoversCurrent",
			sets: []dataSet{
				{interval: interval1, blob: blob1},
				{interval: interval8, blob: blob2},
			},
			wants: []dataSet{
				{interval: interval8, blob: blob2},
			},
		},
		{
			name: "TwoBlobs/OverlappingStart/PortionOfStart",
			sets: []dataSet{
				{interval: interval1, blob: blob1},
				{interval: interval9, blob: blob2},
			},
			wants: []dataSet{
				{interval: interval9, blob: blob2},
				{interval: i9split, blob: blob1},
			},
		},
		{
			name: "TwoBlobs/OverlappingStart/MatchingEnds",
			sets: []dataSet{
				{interval: interval1, blob: blob1},
				{interval: interval10, blob: blob2},
			},
			wants: []dataSet{
				{interval: interval10, blob: blob2},
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			tf := mapping.NewTimeframeReplacer[int, blob]()

			for i := range testcase.sets {
				_ = tf.Add(testcase.sets[i].interval, map[int]blob{testcase.sets[i].blob.user.id: testcase.sets[i].blob})
			}

			newTF, err := tf.Organize()
			require.NoError(t, err)

			seq := newTF.All()

			require.True(t, seq(verifySeqKV(testcase.wants)))
		})
	}
}

func verifySeqKV(wants []dataSet) func(interval mapping.Interval, m map[int]blob) bool {
	return func(interval mapping.Interval, m map[int]blob) bool {
		if m == nil {
			return false
		}

		idx := slices.IndexFunc(wants, func(set dataSet) bool {
			return set.interval == interval
		})

		if idx < 0 {
			return false
		}

		v, ok := m[wants[idx].blob.user.id]
		if !ok {
			return false
		}

		return v == wants[idx].blob
	}
}
