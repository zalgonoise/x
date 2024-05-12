package mapping

import (
	"errors"
	"testing"
	"time"
)

func TestTimeframe(t *testing.T) {
	interval1 := Interval{
		From: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}
	interval2 := Interval{
		From: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
	}

	// interleaved on tail
	interval3 := Interval{
		From: time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
	}
	interval3Split1 := Interval{
		From: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC),
	}
	interval3Split2 := Interval{
		From: time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	// interleaved on head
	interval4 := Interval{
		From: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC),
	}

	interval4Split1 := Interval{
		From: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	interval4Split2 := Interval{
		From: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC),
	}
	interval4Split3 := Interval{
		From: time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	// interleaved in the middle
	interval5 := Interval{
		From: time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 16, 0, 0, 0, time.UTC),
	}

	interval5Split1 := Interval{
		From: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
	}
	interval5Split2 := Interval{
		From: time.Date(2024, 1, 1, 16, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	// interleaved recursively
	interval6 := Interval{
		From: time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC),
	}
	interval7 := Interval{
		From: time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 22, 0, 0, 0, time.UTC),
	}
	interval8 := Interval{
		From: time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 21, 0, 0, 0, time.UTC),
	}

	interval6Split1 := Interval{
		From: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
	}
	interval6Split2 := Interval{
		From: time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC),
	}
	interval6Split3 := Interval{
		From: time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
	}

	interval6Split4 := Interval{
		From: time.Date(2024, 1, 1, 21, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 22, 0, 0, 0, time.UTC),
	}
	interval6Split5 := Interval{
		From: time.Date(2024, 1, 1, 22, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC),
	}
	interval6Split6 := Interval{
		From: time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	kv1 := map[string]string{
		"a": "value",
		"b": "value",
	}

	kv2 := map[string]string{
		"c": "value",
		"d": "value",
	}

	kv3 := map[string]string{"c": "value"}
	kv4 := map[string]string{"d": "value"}
	kv5 := map[string]string{"e": "value"}

	for _, testcase := range []struct {
		name  string
		input map[Interval]map[string]string
		print []Interval
	}{
		{
			name: "sequential",
			input: map[Interval]map[string]string{
				interval1: kv1,
				interval2: kv2,
			},
			print: []Interval{interval1, interval2},
		},
		{
			name: "interleaved/on_tail",
			input: map[Interval]map[string]string{
				interval1: kv1,
				interval3: kv2,
			},
			print: []Interval{interval3Split1, interval3Split2, interval2},
		},
		{
			name: "interleaved/on_head",
			input: map[Interval]map[string]string{
				interval1: kv1,
				interval4: kv2,
			},
			print: []Interval{interval4Split1, interval4Split2, interval4Split3},
		},
		{
			name: "interleaved/on_middle",
			input: map[Interval]map[string]string{
				interval1: kv1,
				interval5: kv2,
			},
			print: []Interval{interval5Split1, interval5, interval5Split2},
		},
		{
			name: "interleaved/recursive",
			input: map[Interval]map[string]string{
				interval1: kv1,
				interval6: kv3,
				interval7: kv4,
				interval8: kv5,
			},
			print: []Interval{
				interval6Split1, interval6Split2, interval6Split3, interval8, interval6Split4, interval6Split5, interval6Split6,
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			tf := NewTimeframe[string, string]()

			for interval, values := range testcase.input {
				ok := tf.Add(interval, values)
				isEqual(t, true, ok)
			}

			tf, err := tf.Organize(func(a, b string) bool {
				return a == b
			})
			if err != nil {
				t.Error(err)
				t.Fail()
			}

			for i := range testcase.print {
				itf, ok := tf.Index.values[testcase.print[i]]
				isEqual(t, true, ok)

				t.Log(itf)
			}
		})
	}
}

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
		err      error
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
			sets, overlaps, err := replace(testcase.cur, testcase.next)
			if err != nil {
				isEqual(t, true, errors.Is(err, testcase.err))
			}

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
		err      error
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
			sets, overlaps, err := split(testcase.cur, testcase.next)
			if err != nil {
				isEqual(t, true, errors.Is(err, testcase.err))
			}

			isEqual(t, testcase.overlaps, overlaps)
			isEqual(t, len(testcase.wants), len(sets))

			for i, w := range testcase.wants {
				isEqual(t, w, sets[i])
			}
		})
	}
}

func FuzzSplit(f *testing.F) {
	f.Add(
		time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC).Unix(),
		time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC).Unix(),
		time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC).Unix(),
		time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC).Unix(),
	)

	f.Fuzz(func(t *testing.T, aFrom, aTo, bFrom, bTo int64) {
		interval1 := Interval{
			From: time.Unix(aFrom, 0),
			To:   time.Unix(aTo, 0),
		}
		interval2 := Interval{
			From: time.Unix(bFrom, 0),
			To:   time.Unix(bTo, 0),
		}

		_, _, err := split(interval1, interval2)
		if err != nil {
			t.Error(err)
		}
	})
}

func FuzzReplace(f *testing.F) {
	f.Add(
		time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC).Unix(),
		time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC).Unix(),
		time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC).Unix(),
		time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC).Unix(),
	)

	f.Fuzz(func(t *testing.T, aFrom, aTo, bFrom, bTo int64) {
		interval1 := Interval{
			From: time.Unix(aFrom, 0),
			To:   time.Unix(aTo, 0),
		}
		interval2 := Interval{
			From: time.Unix(bFrom, 0),
			To:   time.Unix(bTo, 0),
		}

		_, _, err := replace(interval1, interval2)
		if err != nil {
			t.Error(err)
		}
	})
}

func TestCoalesce(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		input map[string]string
		next  map[string]string
		wants map[string]string
	}{
		{
			name:  "simple",
			input: map[string]string{"a": "value"},
			next:  map[string]string{"b": "value"},
			wants: map[string]string{"a": "value", "b": "value"},
		},
		{
			name:  "overwrite",
			input: map[string]string{"a": "value"},
			next:  map[string]string{"a": "value2"},
			wants: map[string]string{"a": "value2"},
		},
		{
			name:  "next_is_nil",
			input: map[string]string{"a": "value"},
			wants: map[string]string{"a": "value"},
		},
		{
			name:  "start_is_nil",
			next:  map[string]string{"a": "value"},
			wants: map[string]string{"a": "value"},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			i := coalesce(testcase.input, testcase.next)

			for k, v := range testcase.wants {
				value, ok := i[k]

				isEqual(t, true, ok)
				isEqual(t, v, value)
			}
		})
	}
}

func TestCoalesceUnset(t *testing.T) {
	for _, testcase := range []struct {
		name    string
		input   map[string]string
		next    map[string]string
		wants   map[string]string
		skipped []string
	}{
		{
			name:  "simple",
			input: map[string]string{"a": "value"},
			next:  map[string]string{"b": "value"},
			wants: map[string]string{"a": "value", "b": "value"},
		},
		{
			name:    "overwrite",
			input:   map[string]string{"a": "value"},
			next:    map[string]string{"a": "value2"},
			wants:   map[string]string{"a": "value"},
			skipped: []string{"a"},
		},
		{
			name:  "next_is_nil",
			input: map[string]string{"a": "value"},
			wants: map[string]string{"a": "value"},
		},
		{
			name:  "start_is_nil",
			next:  map[string]string{"a": "value"},
			wants: map[string]string{"a": "value"},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			i, skipped := coalesceUnset(testcase.input, testcase.next)

			for k, v := range testcase.wants {
				value, ok := i[k]

				isEqual(t, true, ok)
				isEqual(t, v, value)
			}

			isEqual(t, len(testcase.skipped), len(skipped))
			for idx := range testcase.skipped {
				isEqual(t, testcase.skipped[idx], skipped[idx])
			}
		})
	}
}
