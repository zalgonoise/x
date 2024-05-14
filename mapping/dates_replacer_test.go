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

func (b blob) Name() string {
	return b.user.name
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
		sets  []mapping.DataInterval[blob]
		wants []mapping.DataInterval[blob]
	}{
		{
			name:  "OneBlob",
			sets:  []mapping.DataInterval[blob]{{Interval: interval1, Data: blob1}},
			wants: []mapping.DataInterval[blob]{{Interval: interval1, Data: blob1}},
		},
		{
			name: "TwoBlobs/Separate/NextIsAfter",
			sets: []mapping.DataInterval[blob]{
				{Interval: interval1, Data: blob1},
				{Interval: interval2, Data: blob2},
			},
			wants: []mapping.DataInterval[blob]{
				{Interval: interval1, Data: blob1},
				{Interval: interval2, Data: blob2},
			},
		},
		{
			name: "TwoBlobs/Separate/NextIsBefore",
			sets: []mapping.DataInterval[blob]{
				{Interval: interval2, Data: blob2},
				{Interval: interval1, Data: blob1},
			},
			wants: []mapping.DataInterval[blob]{
				{Interval: interval1, Data: blob1},
				{Interval: interval2, Data: blob2},
			},
		},
		{
			name: "TwoBlobs/MatchingStart/NextOverlapsCurrent",
			sets: []mapping.DataInterval[blob]{
				{Interval: interval1, Data: blob1},
				{Interval: interval3, Data: blob2},
			},
			wants: []mapping.DataInterval[blob]{
				{Interval: interval3, Data: blob2},
			},
		},
		{
			name: "TwoBlobs/MatchingStart/NextOverlapsCurrent",
			sets: []mapping.DataInterval[blob]{
				{Interval: interval1, Data: blob1},
				{Interval: interval3, Data: blob2},
			},
			wants: []mapping.DataInterval[blob]{
				{Interval: interval3, Data: blob2},
			},
		},
		{
			name: "TwoBlobs/MatchingStart/NextWithinCurrent",
			sets: []mapping.DataInterval[blob]{
				{Interval: interval1, Data: blob1},
				{Interval: interval4, Data: blob2},
			},
			wants: []mapping.DataInterval[blob]{
				{Interval: interval4, Data: blob2},
				{Interval: i4split, Data: blob1},
			},
		},
		{
			name: "TwoBlobs/MatchingStart/NextIsSameRange",
			sets: []mapping.DataInterval[blob]{
				{Interval: interval1, Data: blob1},
				{Interval: interval1, Data: blob2},
			},
			wants: []mapping.DataInterval[blob]{
				{Interval: interval1, Data: blob2},
			},
		},
		{
			name: "TwoBlobs/OverlappingMiddle/NextWithinCurrent",
			sets: []mapping.DataInterval[blob]{
				{Interval: interval1, Data: blob1},
				{Interval: interval5, Data: blob2},
			},
			wants: []mapping.DataInterval[blob]{
				{Interval: i5split1, Data: blob1},
				{Interval: interval5, Data: blob2},
				{Interval: i5split2, Data: blob1},
			},
		},
		{
			name: "TwoBlobs/OverlappingEnd/NextGoesBeyondCurrent",
			sets: []mapping.DataInterval[blob]{
				{Interval: interval1, Data: blob1},
				{Interval: interval6, Data: blob2},
			},
			wants: []mapping.DataInterval[blob]{
				{Interval: i6split, Data: blob1},
				{Interval: interval6, Data: blob2},
			},
		},
		{
			name: "TwoBlobs/OverlappingEnd/NextMatchesEnds",
			sets: []mapping.DataInterval[blob]{
				{Interval: interval1, Data: blob1},
				{Interval: interval7, Data: blob2},
			},
			wants: []mapping.DataInterval[blob]{
				{Interval: i7split, Data: blob1},
				{Interval: interval7, Data: blob2},
			},
		},
		{
			name: "TwoBlobs/OverlappingStart/NextCoversCurrent",
			sets: []mapping.DataInterval[blob]{
				{Interval: interval1, Data: blob1},
				{Interval: interval8, Data: blob2},
			},
			wants: []mapping.DataInterval[blob]{
				{Interval: interval8, Data: blob2},
			},
		},
		{
			name: "TwoBlobs/OverlappingStart/PortionOfStart",
			sets: []mapping.DataInterval[blob]{
				{Interval: interval1, Data: blob1},
				{Interval: interval9, Data: blob2},
			},
			wants: []mapping.DataInterval[blob]{
				{Interval: interval9, Data: blob2},
				{Interval: i9split, Data: blob1},
			},
		},
		{
			name: "TwoBlobs/OverlappingStart/MatchingEnds",
			sets: []mapping.DataInterval[blob]{
				{Interval: interval1, Data: blob1},
				{Interval: interval10, Data: blob2},
			},
			wants: []mapping.DataInterval[blob]{
				{Interval: interval10, Data: blob2},
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			tf := mapping.NewTimeframeReplacer[int, blob]()

			for i := range testcase.sets {
				_ = tf.Add(testcase.sets[i].Interval, map[int]blob{testcase.sets[i].Data.user.id: testcase.sets[i].Data})
			}

			newTF := tf.Organize(0)

			seq := newTF.All()

			require.True(t, seq(verifySeqKV(testcase.wants)))
		})

		t.Run("OrganizeTimeframeRange/"+testcase.name, func(t *testing.T) {
			var fn = func(yield func(mapping.Interval, map[int]blob) bool) bool {
				for i := range testcase.sets {
					if !yield(testcase.sets[i].Interval, map[int]blob{
						testcase.sets[i].Data.user.id: testcase.sets[i].Data,
					}) {
						return false
					}
				}

				return true
			}

			tf := mapping.Organize[*mapping.TimeframeReplacer[int, blob]](fn, mapping.Replace[map[int]blob](0))

			require.True(t, tf.All()(verifySeqKV(testcase.wants)))
		})
	}
}

func verifySeqKV(wants []mapping.DataInterval[blob]) func(interval mapping.Interval, m map[int]blob) bool {
	return func(interval mapping.Interval, m map[int]blob) bool {
		if m == nil {
			return false
		}

		idx := slices.IndexFunc(wants, func(set mapping.DataInterval[blob]) bool {
			return set.Interval == interval
		})

		if idx < 0 {
			return false
		}

		v, ok := m[wants[idx].Data.user.id]
		if !ok {
			return false
		}

		return v == wants[idx].Data
	}
}
