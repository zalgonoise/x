package mapping_test

import (
	"testing"
	"time"

	"github.com/zalgonoise/x/mapping"
)

func FuzzOrganize(f *testing.F) {
	f.Add(
		int64(1704110400), int64(1704153599), 1,
		int64(1704114000), int64(1704149999), 2,
		int64(1704120000), int64(1704160000), 3,
	)
	f.Add(
		int64(100), int64(500), 10,
		int64(300), int64(700), 10,
		int64(200), int64(600), 20,
	)

	f.Fuzz(func(
		t *testing.T,
		aFrom, aTo int64, aVal int,
		bFrom, bTo int64, bVal int,
		cFrom, cTo int64, cVal int,
	) {
		// Bound timestamps to a reasonable range to avoid time.Time overflow when adding offsets.
		const maxUnix = 1 << 40
		for _, v := range []int64{aFrom, aTo, bFrom, bTo, cFrom, cTo} {
			if v < -maxUnix || v > maxUnix {
				return
			}
		}

		input := []mapping.DataInterval[int]{
			{Interval: mapping.Interval{From: time.Unix(aFrom, 0).UTC(), To: time.Unix(aTo, 0).UTC()}, Data: aVal},
			{Interval: mapping.Interval{From: time.Unix(bFrom, 0).UTC(), To: time.Unix(bTo, 0).UTC()}, Data: bVal},
			{Interval: mapping.Interval{From: time.Unix(cFrom, 0).UTC(), To: time.Unix(cTo, 0).UTC()}, Data: cVal},
		}

		for _, offset := range []time.Duration{0, time.Second} {
			for _, res := range [][]mapping.DataInterval[int]{
				mapping.Organize(input, mapping.WithReplace[int](), mapping.WithOffset[int](offset)),
				mapping.Organize(input, mapping.WithKeep[int](), mapping.WithOffset[int](offset)),
				mapping.Organize(input, mapping.WithFlatten(func(a, b int) int { return a + b }), mapping.WithOffset[int](offset)),
				mapping.Organize(input, mapping.WithIntersect(func(a, b int) int { return a + b }), mapping.WithOffset[int](offset)),
			} {
				verifyDisjointInvariants(t, res, offset)
			}
		}

		appended := mapping.Organize(input, mapping.WithAppend[int]())
		for i := range appended {
			if !appended[i].Interval.IsValid() || appended[i].Interval.From.Equal(appended[i].Interval.To) {
				t.Fatalf("invalid interval in WithAppend result: %+v", appended[i].Interval)
			}
			if i > 0 && appended[i-1].Interval.Compare(appended[i].Interval) > 0 {
				t.Fatalf("WithAppend output not sorted at %d: %+v > %+v", i, appended[i-1].Interval, appended[i].Interval)
			}
		}
	})
}

func verifyDisjointInvariants(t *testing.T, res []mapping.DataInterval[int], offset time.Duration) {
	t.Helper()

	for i := range res {
		cur := res[i].Interval
		if offset == 0 && !cur.From.Before(cur.To) {
			t.Fatalf("expected From < To when offset == 0, got %+v", cur)
		}
		if offset > 0 && cur.To.Before(cur.From) {
			t.Fatalf("expected From <= To when offset > 0, got %+v", cur)
		}

		if i == 0 {
			continue
		}

		prev := res[i-1].Interval
		if offset == 0 && prev.To.After(cur.From) {
			t.Fatalf("overlapping intervals in output at %d: %+v and %+v", i, prev, cur)
		}
		if offset > 0 && !prev.To.Before(cur.From) {
			t.Fatalf("overlapping inclusive intervals in output at %d: %+v and %+v", i, prev, cur)
		}

		if cur.From.Sub(prev.To) <= offset && res[i-1].Data == res[i].Data {
			t.Fatalf("uncoalesced adjacent intervals with equal data at %d: %+v and %+v", i, res[i-1], res[i])
		}
	}
}
