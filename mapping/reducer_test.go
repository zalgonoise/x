package mapping_test

import (
	"testing"
	"time"

	"github.com/zalgonoise/x/mapping"
)

func TestStrategies(t *testing.T) {
	t10 := time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC)
	t12 := time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC)
	t14 := time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC)
	t16 := time.Date(2024, 1, 10, 16, 0, 0, 0, time.UTC)
	t18 := time.Date(2024, 1, 10, 18, 0, 0, 0, time.UTC)

	input := []mapping.DataInterval[string]{
		{Interval: mapping.Interval{From: t10, To: t14}, Data: "A"},
		{Interval: mapping.Interval{From: t12, To: t16}, Data: "B"},
	}

	t.Run("WithReplace", func(t *testing.T) {
		got := mapping.Organize(input, mapping.WithReplace[string]())
		wants := []mapping.DataInterval[string]{
			{Interval: mapping.Interval{From: t10, To: t12}, Data: "A"},
			{Interval: mapping.Interval{From: t12, To: t16}, Data: "B"},
		}
		assertIntervalsEqual(t, wants, got)

		// Also verify passing MergeReplace to WithMergeFunc produces identical results.
		gotMerge := mapping.Organize(input, mapping.WithMergeFunc(mapping.MergeReplace[string]))
		assertIntervalsEqual(t, wants, gotMerge)
	})

	t.Run("WithKeep", func(t *testing.T) {
		got := mapping.Organize(input, mapping.WithKeep[string]())
		wants := []mapping.DataInterval[string]{
			{Interval: mapping.Interval{From: t10, To: t14}, Data: "A"},
			{Interval: mapping.Interval{From: t14, To: t16}, Data: "B"},
		}
		assertIntervalsEqual(t, wants, got)

		// Also verify passing MergeKeep to WithMergeFunc produces identical results.
		gotMerge := mapping.Organize(input, mapping.WithMergeFunc(mapping.MergeKeep[string]))
		assertIntervalsEqual(t, wants, gotMerge)
	})

	t.Run("WithFlatten", func(t *testing.T) {
		got := mapping.Organize(input, mapping.WithFlatten(func(cur, next string) string {
			return cur + "+" + next
		}))
		wants := []mapping.DataInterval[string]{
			{Interval: mapping.Interval{From: t10, To: t12}, Data: "A"},
			{Interval: mapping.Interval{From: t12, To: t14}, Data: "A+B"},
			{Interval: mapping.Interval{From: t14, To: t16}, Data: "B"},
		}
		assertIntervalsEqual(t, wants, got)
	})

	t.Run("WithIntersect/Overlapping", func(t *testing.T) {
		got := mapping.Organize(input, mapping.WithIntersect(func(cur, next string) string {
			return cur + "&" + next
		}))
		wants := []mapping.DataInterval[string]{
			{Interval: mapping.Interval{From: t12, To: t14}, Data: "A&B"},
		}
		assertIntervalsEqual(t, wants, got)
	})

	t.Run("WithIntersect/DisjointBecomesEmpty", func(t *testing.T) {
		disjoint := []mapping.DataInterval[string]{
			{Interval: mapping.Interval{From: t10, To: t12}, Data: "A"},
			{Interval: mapping.Interval{From: t14, To: t16}, Data: "B"},
			{Interval: mapping.Interval{From: t10, To: t18}, Data: "C"},
		}
		got := mapping.Organize(disjoint, mapping.WithIntersect[string](nil))
		if len(got) != 0 {
			t.Fatalf("expected empty intersection for disjoint intervals, got %+v", got)
		}
	})

	t.Run("WithAppend", func(t *testing.T) {
		unordered := []mapping.DataInterval[string]{
			{Interval: mapping.Interval{From: t12, To: t16}, Data: "B"},
			{Interval: mapping.Interval{From: t10, To: t14}, Data: "A1"},
			{Interval: mapping.Interval{From: t10, To: t14}, Data: "A2"},
		}
		got := mapping.Organize(unordered, mapping.WithAppend[string]())
		wants := []mapping.DataInterval[string]{
			{Interval: mapping.Interval{From: t10, To: t14}, Data: "A1"},
			{Interval: mapping.Interval{From: t10, To: t14}, Data: "A2"},
			{Interval: mapping.Interval{From: t12, To: t16}, Data: "B"},
		}
		assertIntervalsEqual(t, wants, got)

		// Also verify TimeframeStream with WithAppend.
		stream := mapping.NewTimeframeStream(mapping.WithAppend[string]())
		stream.AddSlice(unordered...)
		assertIntervalsEqual(t, wants, stream.Slice())

		// Also verify OrganizeSeq with WithAppend.
		seqGot := mapping.Collect(mapping.OrganizeSeq(mapping.AsSeq(unordered), mapping.WithAppend[string]()))
		assertIntervalsEqual(t, wants, seqGot)
	})
}

func TestWithOffset(t *testing.T) {
	t.Run("MiddleSplitWith1SecondOffset", func(t *testing.T) {
		input := []mapping.DataInterval[string]{
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 18, 0, 0, 0, time.UTC),
				},
				Data: "A",
			},
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 16, 0, 0, 0, time.UTC),
				},
				Data: "B",
			},
		}

		got := mapping.Organize(
			input,
			mapping.WithFlatten(func(cur, next string) string { return cur + next }),
			mapping.WithOffset[string](time.Second),
		)

		wants := []mapping.DataInterval[string]{
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 13, 59, 59, 0, time.UTC),
				},
				Data: "A",
			},
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 16, 0, 0, 0, time.UTC),
				},
				Data: "AB",
			},
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 16, 0, 1, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 18, 0, 0, 0, time.UTC),
				},
				Data: "A",
			},
		}

		assertIntervalsEqual(t, wants, got)
	})

	t.Run("MultiConflictWithGapAndOffset", func(t *testing.T) {
		input := []mapping.DataInterval[string]{
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
				},
				Data: "A",
			},
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
				},
				Data: "B",
			},
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 9, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 13, 0, 0, 0, time.UTC),
				},
				Data: "C",
			},
		}

		got := mapping.Organize(
			input,
			mapping.WithFlatten(func(cur, next string) string { return cur + next }),
			mapping.WithOffset[string](time.Second),
		)

		wants := []mapping.DataInterval[string]{
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 8, 59, 59, 0, time.UTC),
				},
				Data: "A",
			},
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 9, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
				},
				Data: "AC",
			},
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 10, 0, 1, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 11, 59, 59, 0, time.UTC),
				},
				Data: "C",
			},
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 13, 0, 0, 0, time.UTC),
				},
				Data: "BC",
			},
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 13, 0, 1, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
				},
				Data: "B",
			},
		}

		assertIntervalsEqual(t, wants, got)
	})

	t.Run("CoalesceAdjacentSeparatedByOffset", func(t *testing.T) {
		input := []mapping.DataInterval[string]{
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 9, 59, 59, 0, time.UTC),
				},
				Data: "A",
			},
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
				},
				Data: "A",
			},
		}

		got := mapping.Organize(input, mapping.WithOffset[string](time.Second))
		wants := []mapping.DataInterval[string]{
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
				},
				Data: "A",
			},
		}

		assertIntervalsEqual(t, wants, got)
	})
}

func TestRegressions(t *testing.T) {
	t.Run("DisjointIntervalsWithEqualDataMustNotCoalesceAcrossGap", func(t *testing.T) {
		// Regression test: previously mergeCache checked prev.To.Sub(cur.From) <= offset,
		// which was negative (-4h <= 0) when cur.From > prev.To, erroneously merging
		// disjoint intervals with equal Data across the gap.
		i1 := mapping.Interval{
			From: time.Date(2024, 1, 10, 6, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC),
		}
		i2 := mapping.Interval{
			From: time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
		}

		input := []mapping.DataInterval[string]{
			{Interval: i1, Data: "same"},
			{Interval: i2, Data: "same"},
		}

		got := mapping.Organize(input)
		assertIntervalsEqual(t, input, got)

		// Now bridge the exact gap [08:00, 12:00] with the same value: all three should coalesce into [06:00, 14:00].
		bridged := append(input, mapping.DataInterval[string]{
			Interval: mapping.Interval{
				From: time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC),
				To:   time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
			},
			Data: "same",
		})
		gotBridged := mapping.Organize(bridged)
		wantsBridged := []mapping.DataInterval[string]{
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 6, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
				},
				Data: "same",
			},
		}
		assertIntervalsEqual(t, wantsBridged, gotBridged)

		// Also test prepending a touching interval [04:00, 06:00] with the same value.
		prepended := []mapping.DataInterval[string]{
			{Interval: i1, Data: "same"},
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 4, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 6, 0, 0, 0, time.UTC),
				},
				Data: "same",
			},
		}
		gotPrepended := mapping.Organize(prepended)
		wantsPrepended := []mapping.DataInterval[string]{
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 4, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC),
				},
				Data: "same",
			},
		}
		assertIntervalsEqual(t, wantsPrepended, gotPrepended)

		// And test replacing a middle conflicting interval with the same value as its left and right neighbors.
		sandwich := []mapping.DataInterval[string]{
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 6, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC),
				},
				Data: "same",
			},
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
				},
				Data: "other",
			},
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
				},
				Data: "same",
			},
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
				},
				Data: "same",
			},
		}
		gotSandwich := mapping.Organize(sandwich, mapping.WithReplace[string]())
		assertIntervalsEqual(t, wantsBridged, gotSandwich)
	})

	t.Run("OutOfOrderDisjointInsertsFollowedBySpanningOverlap", func(t *testing.T) {
		// Regression test: previously non-conflicting inserts were appended without
		// maintaining sorted order in cache, causing a subsequent spanning overlap to
		// split against out-of-order conflicts and produce inverted intervals (From > To).
		input := []mapping.DataInterval[string]{
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 18, 0, 0, 0, time.UTC),
				},
				Data: "B",
			},
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 13, 0, 0, 0, time.UTC),
				},
				Data: "A",
			},
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 11, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 17, 0, 0, 0, time.UTC),
				},
				Data: "C",
			},
		}

		got := mapping.Organize(input, mapping.WithFlatten(func(cur, next string) string {
			return cur + next
		}))

		wants := []mapping.DataInterval[string]{
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 11, 0, 0, 0, time.UTC),
				},
				Data: "A",
			},
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 11, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 13, 0, 0, 0, time.UTC),
				},
				Data: "AC",
			},
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 13, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC),
				},
				Data: "C",
			},
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 17, 0, 0, 0, time.UTC),
				},
				Data: "BC",
			},
			{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 17, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 10, 18, 0, 0, 0, time.UTC),
				},
				Data: "B",
			},
		}

		assertIntervalsEqual(t, wants, got)
	})

	t.Run("MultiConflictSpanning5IntervalsWithGaps", func(t *testing.T) {
		iA := mapping.Interval{
			From: time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC),
			To:   time.Date(2020, 1, 1, 13, 0, 0, 0, time.UTC),
		}
		iB := mapping.Interval{
			From: time.Date(2020, 1, 1, 13, 0, 0, 0, time.UTC),
			To:   time.Date(2020, 1, 1, 14, 0, 0, 0, time.UTC),
		}
		iC := mapping.Interval{
			From: time.Date(2020, 1, 1, 15, 0, 0, 0, time.UTC),
			To:   time.Date(2020, 1, 1, 16, 0, 0, 0, time.UTC),
		}
		iD := mapping.Interval{
			From: time.Date(2020, 1, 1, 17, 0, 0, 0, time.UTC),
			To:   time.Date(2020, 1, 1, 19, 0, 0, 0, time.UTC),
		}
		iE := mapping.Interval{
			From: time.Date(2020, 1, 1, 19, 0, 0, 0, time.UTC),
			To:   time.Date(2020, 1, 1, 22, 0, 0, 0, time.UTC),
		}
		iF := mapping.Interval{
			From: time.Date(2020, 1, 1, 11, 0, 0, 0, time.UTC),
			To:   time.Date(2020, 1, 1, 21, 0, 0, 0, time.UTC),
		}

		input := []mapping.DataInterval[int]{
			{Interval: iA, Data: 1},
			{Interval: iB, Data: 2},
			{Interval: iC, Data: 3},
			{Interval: iD, Data: 4},
			{Interval: iE, Data: 5},
			{Interval: iF, Data: 6},
		}

		got := mapping.Organize(input, mapping.WithReplace[int]())
		wants := []mapping.DataInterval[int]{
			{
				Interval: mapping.Interval{
					From: time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC),
					To:   time.Date(2020, 1, 1, 11, 0, 0, 0, time.UTC),
				},
				Data: 1,
			},
			{
				Interval: iF,
				Data:     6,
			},
			{
				Interval: mapping.Interval{
					From: time.Date(2020, 1, 1, 21, 0, 0, 0, time.UTC),
					To:   time.Date(2020, 1, 1, 22, 0, 0, 0, time.UTC),
				},
				Data: 5,
			},
		}

		assertIntervalsEqual(t, wants, got)
	})
}
