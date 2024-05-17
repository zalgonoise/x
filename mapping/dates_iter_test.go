package mapping

import (
	"testing"
	"time"
)

func TestReplace(t *testing.T) {
	intervalBefore := Interval{
		From: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 11, 59, 59, 0, time.UTC),
	}
	intervalAfter := Interval{
		From: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
	}

	interval1 := Interval{
		From: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
	}
	interval2 := Interval{
		From: time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC),
	}
	interval3 := Interval{
		From: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
	}
	interval4 := Interval{
		From: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
	}
	interval5 := Interval{
		From: time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
	}
	interval6 := Interval{
		From: time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
	}
	interval7 := Interval{
		From: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
	}
	interval8 := Interval{
		From: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 22, 0, 0, 0, time.UTC),
	}
	interval9 := Interval{
		From: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
	}

	i1Split1 := Interval{
		From: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
	}
	i1Split2 := Interval{
		From: time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
	}

	i4split := Interval{
		From: time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
	}
	i5split := Interval{
		From: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
	}
	i6Split := Interval{
		From: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
	}
	i8Split := Interval{
		From: time.Date(2024, 1, 1, 22, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
	}

	for _, testcase := range []struct {
		name     string
		cur      Interval
		next     Interval
		wants    []IntervalSet
		overlaps bool
	}{
		{
			name: "CurBeforeNext",
			cur:  intervalBefore,
			next: interval1,
			wants: []IntervalSet{
				{cur: true, i: intervalBefore},
				{next: true, i: interval1},
			},
		},
		{
			name: "CurAfterNext",
			cur:  intervalAfter,
			next: interval1,
			wants: []IntervalSet{
				{next: true, i: interval1},
				{cur: true, i: intervalAfter},
			},
		},
		{
			name: "OverlappingStart/NextEndsAfter",
			cur:  interval1,
			next: interval3,
			wants: []IntervalSet{
				{next: true, i: interval3},
			},
			overlaps: true,
		},
		{
			name: "OverlappingStart/NextEndsBefore",
			cur:  interval1,
			next: interval4,
			wants: []IntervalSet{
				{next: true, i: interval4},
				{cur: true, i: i4split},
			},
			overlaps: true,
		},
		{
			name: "OverlappingStart/MatchingLengths",
			cur:  interval1,
			next: interval1,
			wants: []IntervalSet{
				{next: true, i: interval1},
			},
			overlaps: true,
		},
		{
			name: "IntersectMiddle/NextEndsAfter",
			cur:  interval1,
			next: interval5,
			wants: []IntervalSet{
				{cur: true, i: i5split},
				{next: true, i: interval5},
			},
			overlaps: true,
		},
		{
			name: "IntersectMiddle/NextIsWithin",
			cur:  interval1,
			next: interval2,
			wants: []IntervalSet{
				{cur: true, i: i1Split1},
				{next: true, i: interval2},
				{cur: true, i: i1Split2},
			},
			overlaps: true,
		},
		{
			name: "IntersectMiddle/NextEndsLikeCur",
			cur:  interval1,
			next: interval6,
			wants: []IntervalSet{
				{cur: true, i: i6Split},
				{next: true, i: interval6},
			},
			overlaps: true,
		},
		{
			name: "IntersectBeginning/NextOverlapsCur",
			cur:  interval1,
			next: interval7,
			wants: []IntervalSet{
				{next: true, i: interval7},
			},
			overlaps: true,
		},
		{
			name: "IntersectBeginning/NextOverlapsStart",
			cur:  interval1,
			next: interval8,
			wants: []IntervalSet{
				{next: true, i: interval8},
				{cur: true, i: i8Split},
			},
			overlaps: true,
		},
		{
			name: "IntersectBeginning/NextEndsWithStart",
			cur:  interval1,
			next: interval9,
			wants: []IntervalSet{
				{next: true, i: interval9},
			},
			overlaps: true,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			sets, overlaps := replace(testcase.cur, testcase.next, 0)

			isEqual(t, testcase.overlaps, overlaps)
			isEqual(t, len(testcase.wants), len(sets))

			for i, w := range testcase.wants {
				isEqual(t, w, sets[i])
			}
		})
	}
}

func TestSplit(t *testing.T) {
	intervalBefore := Interval{
		From: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 11, 59, 59, 0, time.UTC),
	}
	intervalAfter := Interval{
		From: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
	}

	interval1 := Interval{
		From: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
	}
	interval2 := Interval{
		From: time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC),
	}
	interval3 := Interval{
		From: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
	}
	interval4 := Interval{
		From: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
	}
	interval5 := Interval{
		From: time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
	}
	interval6 := Interval{
		From: time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
	}
	interval7 := Interval{
		From: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
	}
	interval8 := Interval{
		From: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 22, 0, 0, 0, time.UTC),
	}
	interval9 := Interval{
		From: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
	}

	i1Split1 := Interval{
		From: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
	}
	i1Split2 := Interval{
		From: time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
	}
	i3split := Interval{
		From: time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
		To:   time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
	}
	i4split := Interval{
		From: time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
	}
	i5split1 := Interval{
		From: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
	}
	i5split2 := Interval{
		From: time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
	}
	i5split3 := Interval{
		From: time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
		To:   time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
	}
	i6Split := Interval{
		From: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
	}
	i7Split1 := Interval{
		From: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	i7Split2 := Interval{
		From: time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
		To:   time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
	}
	i8Split1 := Interval{
		From: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	i8Split2 := Interval{
		From: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 22, 0, 0, 0, time.UTC),
	}
	i8Split3 := Interval{
		From: time.Date(2024, 1, 1, 22, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
	}
	i9Split := Interval{
		From: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	for _, testcase := range []struct {
		name     string
		cur      Interval
		next     Interval
		wants    []IntervalSet
		overlaps bool
	}{
		{
			name: "CurBeforeNext",
			cur:  intervalBefore,
			next: interval1,
			wants: []IntervalSet{
				{cur: true, i: intervalBefore},
				{next: true, i: interval1},
			},
		},
		{
			name: "CurAfterNext",
			cur:  intervalAfter,
			next: interval1,
			wants: []IntervalSet{
				{next: true, i: interval1},
				{cur: true, i: intervalAfter},
			},
		},
		{
			name: "OverlappingStart/NextEndsAfter",
			cur:  interval1,
			next: interval3,
			wants: []IntervalSet{
				{cur: true, next: true, i: interval1},
				{next: true, i: i3split},
			},
			overlaps: true,
		},
		{
			name: "OverlappingStart/NextEndsBefore",
			cur:  interval1,
			next: interval4,
			wants: []IntervalSet{
				{cur: true, next: true, i: interval4},
				{cur: true, i: i4split},
			},
			overlaps: true,
		},
		{
			name: "OverlappingStart/MatchingLengths",
			cur:  interval1,
			next: interval1,
			wants: []IntervalSet{
				{cur: true, next: true, i: interval1},
			},
			overlaps: true,
		},
		{
			name: "IntersectMiddle/NextEndsAfter",
			cur:  interval1,
			next: interval5,
			wants: []IntervalSet{
				{cur: true, i: i5split1},
				{cur: true, next: true, i: i5split2},
				{next: true, i: i5split3},
			},
			overlaps: true,
		},
		{
			name: "IntersectMiddle/NextIsWithin",
			cur:  interval1,
			next: interval2,
			wants: []IntervalSet{
				{cur: true, i: i1Split1},
				{cur: true, next: true, i: interval2},
				{cur: true, i: i1Split2},
			},
			overlaps: true,
		},
		{
			name: "IntersectMiddle/NextEndsLikeCur",
			cur:  interval1,
			next: interval6,
			wants: []IntervalSet{
				{cur: true, i: i6Split},
				{cur: true, next: true, i: interval6},
			},
			overlaps: true,
		},
		{
			name: "IntersectBeginning/NextOverlapsCur",
			cur:  interval1,
			next: interval7,
			wants: []IntervalSet{
				{next: true, i: i7Split1},
				{cur: true, next: true, i: interval1},
				{next: true, i: i7Split2},
			},
			overlaps: true,
		},
		{
			name: "IntersectBeginning/NextOverlapsStart",
			cur:  interval1,
			next: interval8,
			wants: []IntervalSet{
				{next: true, i: i8Split1},
				{cur: true, next: true, i: i8Split2},
				{cur: true, i: i8Split3},
			},
			overlaps: true,
		},
		{
			name: "IntersectBeginning/NextEndsWithStart",
			cur:  interval1,
			next: interval9,
			wants: []IntervalSet{
				{next: true, i: i9Split},
				{cur: true, next: true, i: interval1},
			},
			overlaps: true,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			sets, overlaps := split(testcase.cur, testcase.next, 0)

			isEqual(t, testcase.overlaps, overlaps)
			isEqual(t, len(testcase.wants), len(sets))

			for i, w := range testcase.wants {
				isEqual(t, w, sets[i])
			}
		})
	}
}
