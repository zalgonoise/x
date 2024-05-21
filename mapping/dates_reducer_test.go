package mapping

import (
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMergeCache(t *testing.T) {
	type blob struct {
		name string
		id   int
	}

	i1 := Interval{
		From: time.Date(2020, 1, 1, 6, 0, 0, 0, time.UTC),
		To:   time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	i2 := Interval{
		From: time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2020, 1, 1, 18, 0, 0, 0, time.UTC),
	}

	i3 := Interval{
		From: time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC),
		To:   time.Date(2020, 1, 1, 16, 0, 0, 0, time.UTC),
	}

	i2Merged := Interval{
		From: time.Date(2020, 1, 1, 6, 0, 0, 0, time.UTC),
		To:   time.Date(2020, 1, 1, 18, 0, 0, 0, time.UTC),
	}

	for _, testcase := range []struct {
		name  string
		cache []DataInterval[blob]
		wants []DataInterval[blob]
	}{
		{
			name:  "NothingToMerge",
			cache: []DataInterval[blob]{},
			wants: []DataInterval[blob]{},
		},
		{
			name: "NothingToMerge/WithItems",
			cache: []DataInterval[blob]{
				{Interval: i1, Data: blob{"blob-a", 1}},
				{Interval: i2, Data: blob{"blob-b", 2}},
			},
			wants: []DataInterval[blob]{
				{Interval: i1, Data: blob{"blob-a", 1}},
				{Interval: i2, Data: blob{"blob-b", 2}},
			},
		},
		{
			name: "MergeHeadAndTail",
			cache: []DataInterval[blob]{
				{Interval: i1, Data: blob{"blob-a", 1}},
				{Interval: i2, Data: blob{"blob-a", 1}},
			},
			wants: []DataInterval[blob]{
				{Interval: i2Merged, Data: blob{"blob-a", 1}},
			},
		},
		{
			name: "MergeWithin/CurWithinNext",
			cache: []DataInterval[blob]{
				{Interval: i3, Data: blob{"blob-a", 1}},
				{Interval: i2Merged, Data: blob{"blob-a", 1}},
			},
			wants: []DataInterval[blob]{
				{Interval: i2Merged, Data: blob{"blob-a", 1}},
			},
		},
		{
			name: "MergeWithin/NextWithinCur",
			cache: []DataInterval[blob]{
				{Interval: i2Merged, Data: blob{"blob-a", 1}},
				{Interval: i3, Data: blob{"blob-a", 1}},
			},
			wants: []DataInterval[blob]{
				{Interval: i2Merged, Data: blob{"blob-a", 1}},
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			cache := mergeCache(testcase.cache, 0)

			require.Equal(t, len(testcase.wants), len(cache))

			slices.SortFunc(testcase.wants, func(a, b DataInterval[blob]) int {
				return a.Interval.From.Compare(b.Interval.From)
			})
			slices.SortFunc(cache, func(a, b DataInterval[blob]) int {
				return a.Interval.From.Compare(b.Interval.From)
			})

			for i := range testcase.wants {
				require.Equal(t, testcase.wants[i].Interval, cache[i].Interval)
				require.Equal(t, testcase.wants[i].Data, cache[i].Data)
			}
		})
	}
}

func TestResolveConflicts(t *testing.T) {
	type blob struct {
		name string
		id   int
	}

	i1 := Interval{
		From: time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2020, 1, 1, 15, 0, 0, 0, time.UTC),
	}
	i2 := Interval{
		From: time.Date(2020, 1, 1, 15, 0, 0, 0, time.UTC),
		To:   time.Date(2020, 1, 1, 18, 0, 0, 0, time.UTC),
	}
	i3 := Interval{
		From: time.Date(2020, 1, 1, 13, 0, 0, 0, time.UTC),
		To:   time.Date(2020, 1, 1, 14, 0, 0, 0, time.UTC),
	}

	i3split1 := Interval{
		From: time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2020, 1, 1, 13, 0, 0, 0, time.UTC),
	}
	i3split2 := Interval{
		From: time.Date(2020, 1, 1, 14, 0, 0, 0, time.UTC),
		To:   time.Date(2020, 1, 1, 15, 0, 0, 0, time.UTC),
	}

	i4 := Interval{
		From: time.Date(2020, 1, 1, 13, 0, 0, 0, time.UTC),
		To:   time.Date(2020, 1, 1, 16, 0, 0, 0, time.UTC),
	}
	i4split1 := Interval{
		From: time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2020, 1, 1, 13, 0, 0, 0, time.UTC),
	}
	i4split2 := Interval{
		From: time.Date(2020, 1, 1, 16, 0, 0, 0, time.UTC),
		To:   time.Date(2020, 1, 1, 18, 0, 0, 0, time.UTC),
	}

	b1 := blob{name: "blob-1", id: 0}
	b2 := blob{name: "blob-2", id: 1}
	b3 := blob{name: "blob-3", id: 2}
	b4 := blob{name: "blob-4", id: 3}
	b5 := blob{name: "blob-5", id: 4}
	b6 := blob{name: "blob-6", id: 5}

	iA := Interval{
		From: time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC),
		To:   time.Date(2020, 1, 1, 13, 0, 0, 0, time.UTC),
	}
	iB := Interval{
		From: time.Date(2020, 1, 1, 13, 0, 0, 0, time.UTC),
		To:   time.Date(2020, 1, 1, 14, 0, 0, 0, time.UTC),
	}
	iC := Interval{
		From: time.Date(2020, 1, 1, 15, 0, 0, 0, time.UTC),
		To:   time.Date(2020, 1, 1, 16, 0, 0, 0, time.UTC),
	}
	iD := Interval{
		From: time.Date(2020, 1, 1, 17, 0, 0, 0, time.UTC),
		To:   time.Date(2020, 1, 1, 19, 0, 0, 0, time.UTC),
	}
	iE := Interval{
		From: time.Date(2020, 1, 1, 19, 0, 0, 0, time.UTC),
		To:   time.Date(2020, 1, 1, 22, 0, 0, 0, time.UTC),
	}
	iF := Interval{
		From: time.Date(2020, 1, 1, 11, 0, 0, 0, time.UTC),
		To:   time.Date(2020, 1, 1, 21, 0, 0, 0, time.UTC),
	}

	complSplit1 := Interval{
		From: time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC),
		To:   time.Date(2020, 1, 1, 11, 0, 0, 0, time.UTC),
	}
	complSplit2 := Interval{
		From: time.Date(2020, 1, 1, 21, 0, 0, 0, time.UTC),
		To:   time.Date(2020, 1, 1, 22, 0, 0, 0, time.UTC),
	}

	for _, testcase := range []struct {
		name  string
		cache []DataInterval[blob]
		next  DataInterval[blob]
		wants []DataInterval[blob]
	}{
		{
			name:  "EmptyCache",
			cache: make([]DataInterval[blob], 0),
			next:  DataInterval[blob]{Data: b1, Interval: i1},
			wants: []DataInterval[blob]{
				{Data: b1, Interval: i1},
			},
		},
		{
			name: "NoConflict",
			cache: []DataInterval[blob]{
				{Data: b1, Interval: i1},
			},
			next: DataInterval[blob]{Data: b2, Interval: i2},
			wants: []DataInterval[blob]{
				{Data: b1, Interval: i1},
				{Data: b2, Interval: i2},
			},
		},
		{
			name: "WithConflict/SplitInto3",
			cache: []DataInterval[blob]{
				{Data: b1, Interval: i1},
			},
			next: DataInterval[blob]{Data: b2, Interval: i3},
			wants: []DataInterval[blob]{
				{Data: b1, Interval: i3split1},
				{Data: b2, Interval: i3},
				{Data: b1, Interval: i3split2},
			},
		},
		{
			name: "MultiConflict/SplitInto2Each",
			cache: []DataInterval[blob]{
				{Data: b1, Interval: i1},
				{Data: b2, Interval: i2},
			},
			next: DataInterval[blob]{Data: b3, Interval: i4},
			wants: []DataInterval[blob]{
				{Data: b1, Interval: i4split1},
				{Data: b3, Interval: i4},
				{Data: b2, Interval: i4split2},
			},
		},
		{
			name: "MultiConflict/Complex/WithHeadAndTailAndOverlaps",
			cache: []DataInterval[blob]{
				{Data: b1, Interval: iA},
				{Data: b2, Interval: iB},
				{Data: b3, Interval: iC},
				{Data: b4, Interval: iD},
				{Data: b5, Interval: iE},
			},
			next: DataInterval[blob]{Data: b6, Interval: iF},
			wants: []DataInterval[blob]{
				{Data: b1, Interval: complSplit1},
				{Data: b6, Interval: iF},
				{Data: b5, Interval: complSplit2},
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			res := resolveAnyConflicts(
				testcase.next.Interval, testcase.next.Data, testcase.cache,
				split,
				func(a, b blob) bool {
					return a.name == b.name
				},
				func(a, b blob) blob {
					if b.name != "" {
						return b
					}

					return a
				},
				0,
			)

			slices.SortFunc(res, func(a, b DataInterval[blob]) int {
				return a.Interval.From.Compare(b.Interval.From)
			})

			require.Equal(t, len(testcase.wants), len(res))

			for i := range testcase.wants {
				require.Equal(t, testcase.wants[i].Interval, res[i].Interval)
				require.Equal(t, testcase.wants[i].Data, res[i].Data)
			}
		})

	}
}
