package mapping_test

import (
	"maps"
	"slices"
	"testing"
	"time"

	"github.com/zalgonoise/cfg"
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

func TestIntervalMethods(t *testing.T) {
	t0 := time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC)
	t1 := time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 10, 16, 0, 0, 0, time.UTC)

	var zero mapping.Interval
	if !zero.IsZero() {
		t.Fatalf("expected zero interval to report IsZero() == true")
	}
	if zero.IsValid() {
		t.Fatalf("expected zero interval to report IsValid() == false")
	}

	inverted := mapping.Interval{From: t1, To: t0}
	if inverted.IsValid() {
		t.Fatalf("expected inverted interval to report IsValid() == false")
	}
	if inverted.Duration() != 0 {
		t.Fatalf("expected inverted interval duration to be 0, got %v", inverted.Duration())
	}

	valid := mapping.Interval{From: t0, To: t1}
	if !valid.IsValid() {
		t.Fatalf("expected valid interval to report IsValid() == true")
	}
	if valid.Duration() != 2*time.Hour {
		t.Fatalf("expected duration 2h, got %v", valid.Duration())
	}

	if !valid.Contains(t0) || !valid.Contains(t1) || !valid.Contains(t0.Add(time.Hour)) {
		t.Fatalf("expected interval to contain its bounds and midpoint")
	}
	if valid.Contains(t2) {
		t.Fatalf("expected interval not to contain t2")
	}

	other := mapping.Interval{From: t0.Add(time.Hour), To: t2}
	if !valid.Overlaps(other) {
		t.Fatalf("expected overlapping intervals to report Overlaps == true")
	}

	adjacent := mapping.Interval{From: t1, To: t2}
	if valid.Overlaps(adjacent) {
		t.Fatalf("expected touching half-open intervals to report Overlaps == false")
	}

	if valid.Compare(adjacent) != -1 {
		t.Fatalf("expected valid < adjacent")
	}
	if adjacent.Compare(valid) != 1 {
		t.Fatalf("expected adjacent > valid")
	}
	if valid.Compare(valid) != 0 {
		t.Fatalf("expected valid == valid")
	}
	if valid.Compare(mapping.Interval{From: t0, To: t2}) != -1 {
		t.Fatalf("expected tie-break on To to return -1")
	}
}

func TestCollectAndAsSeq(t *testing.T) {
	if got := mapping.Collect[int](nil); got != nil {
		t.Fatalf("expected nil from Collect(nil), got %v", got)
	}

	items := []mapping.DataInterval[int]{
		{
			Interval: mapping.Interval{
				From: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
				To:   time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
			},
			Data: 1,
		},
		{
			Interval: mapping.Interval{
				From: time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
				To:   time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
			},
			Data: 2,
		},
	}

	seq := mapping.AsSeq(items)

	// Test early break in iterator.
	count := 0
	for range seq {
		count++
		break
	}
	if count != 1 {
		t.Fatalf("expected 1 iteration on break, got %d", count)
	}

	collected := mapping.Collect(seq)
	assertIntervalsEqual(t, items, collected)
}

func TestTimeframe(t *testing.T) {
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

	flattenMergeFunc := func(cur, next blob) blob {
		if next.user.name != "" {
			cur.user.name = next.user.name
		}

		return cur
	}

	flattenCmpFunc := func(cur, next blob) bool {
		return cur.user.name == next.user.name
	}

	bA := blob{user: user{id: 1, name: "blob-a"}, data: data{len: 1}}
	bB := blob{user: user{id: 2, name: "blob-b"}, data: data{len: 1}}
	bC := blob{user: user{id: 3, name: "blob-c"}, data: data{len: 1}}
	bD := blob{user: user{id: 4, name: "blob-d"}, data: data{len: 1}}
	bE := blob{user: user{id: 5, name: "blob-e"}, data: data{len: 1}}
	bF := blob{user: user{id: 6, name: "blob-f"}, data: data{len: 1}}

	iA := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 5, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 9, 0, 0, 0, time.UTC),
		},
		Data: bA,
	}
	iB := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 6, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC),
		},
		Data: bB,
	}
	iC := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 7, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC),
		},
		Data: bC,
	}
	iD := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
		},
		Data: bD,
	}
	iE := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 11, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 20, 0, 0, 0, time.UTC),
		},
		Data: bE,
	}
	iF := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 17, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 18, 0, 0, 0, time.UTC),
		},
		Data: bF,
	}

	i3RangesMerged := []mapping.DataInterval[blob]{
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 5, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 6, 0, 0, 0, time.UTC)}, Data: bA},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 6, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 7, 0, 0, 0, time.UTC)}, Data: bB},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 7, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC)}, Data: bC},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)}, Data: bB},
	}

	i3RangesFlattened := []mapping.DataInterval[blob]{
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 5, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 6, 0, 0, 0, time.UTC)}, Data: bA},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 6, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 7, 0, 0, 0, time.UTC)}, Data: blob{user: user{id: 1, name: "blob-b"}, data: data{len: 1}}},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 7, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC)}, Data: blob{user: user{id: 1, name: "blob-c"}, data: data{len: 1}}},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)}, Data: blob{user: user{id: 1, name: "blob-b"}, data: data{len: 1}}},
	}

	i4RangesMerged := []mapping.DataInterval[blob]{
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 5, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 6, 0, 0, 0, time.UTC)}, Data: bA},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 6, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 7, 0, 0, 0, time.UTC)}, Data: bB},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 7, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC)}, Data: bC},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC)}, Data: bB},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC)}, Data: bD},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)}, Data: bB},
	}

	i5RangesMerged := []mapping.DataInterval[blob]{
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 5, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 6, 0, 0, 0, time.UTC)}, Data: bA},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 6, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 7, 0, 0, 0, time.UTC)}, Data: bB},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 7, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC)}, Data: bC},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC)}, Data: bB},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 11, 0, 0, 0, time.UTC)}, Data: bD},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 11, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 20, 0, 0, 0, time.UTC)}, Data: bE},
	}

	i6RangesMerged := []mapping.DataInterval[blob]{
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 5, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 6, 0, 0, 0, time.UTC)}, Data: bA},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 6, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 7, 0, 0, 0, time.UTC)}, Data: bB},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 7, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC)}, Data: bC},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC)}, Data: bB},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 11, 0, 0, 0, time.UTC)}, Data: bD},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 11, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 17, 0, 0, 0, time.UTC)}, Data: bE},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 17, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 18, 0, 0, 0, time.UTC)}, Data: bF},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 18, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 20, 0, 0, 0, time.UTC)}, Data: bE},
	}

	i6WantsFlattened := []mapping.DataInterval[blob]{
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 5, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 6, 0, 0, 0, time.UTC)}, Data: bA},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 6, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 7, 0, 0, 0, time.UTC)}, Data: blob{user: user{id: 1, name: "blob-b"}, data: data{len: 1}}},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 7, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC)}, Data: blob{user: user{id: 1, name: "blob-c"}, data: data{len: 1}}},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC)}, Data: blob{user: user{id: 1, name: "blob-b"}, data: data{len: 1}}},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 11, 0, 0, 0, time.UTC)}, Data: blob{user: user{id: 1, name: "blob-d"}, data: data{len: 1}}},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 11, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 17, 0, 0, 0, time.UTC)}, Data: blob{user: user{id: 1, name: "blob-e"}, data: data{len: 1}}},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 17, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 18, 0, 0, 0, time.UTC)}, Data: blob{user: user{id: 1, name: "blob-f"}, data: data{len: 1}}},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 18, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 20, 0, 0, 0, time.UTC)}, Data: blob{user: user{id: 1, name: "blob-e"}, data: data{len: 1}}},
	}

	iStackA := mapping.DataInterval[blob]{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 23, 0, 0, 0, time.UTC)}, Data: bA}
	iStackB := mapping.DataInterval[blob]{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 13, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 22, 0, 0, 0, time.UTC)}, Data: bB}
	iStackC := mapping.DataInterval[blob]{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 21, 0, 0, 0, time.UTC)}, Data: bC}
	iStackD := mapping.DataInterval[blob]{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 20, 0, 0, 0, time.UTC)}, Data: bD}

	iStackFlattened := []mapping.DataInterval[blob]{
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 13, 0, 0, 0, time.UTC)}, Data: blob{user: user{id: 1, name: "blob-a"}, data: data{len: 1}}},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 13, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC)}, Data: blob{user: user{id: 1, name: "blob-b"}, data: data{len: 1}}},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)}, Data: blob{user: user{id: 1, name: "blob-c"}, data: data{len: 1}}},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 20, 0, 0, 0, time.UTC)}, Data: blob{user: user{id: 1, name: "blob-d"}, data: data{len: 1}}},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 20, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 21, 0, 0, 0, time.UTC)}, Data: blob{user: user{id: 1, name: "blob-c"}, data: data{len: 1}}},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 21, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 22, 0, 0, 0, time.UTC)}, Data: blob{user: user{id: 1, name: "blob-b"}, data: data{len: 1}}},
		{Interval: mapping.Interval{From: time.Date(2024, 1, 10, 22, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 10, 23, 0, 0, 0, time.UTC)}, Data: blob{user: user{id: 1, name: "blob-a"}, data: data{len: 1}}},
	}

	flattenOpts := []cfg.Option[mapping.Config[blob]]{
		mapping.WithFlatten(flattenMergeFunc),
		mapping.WithCompareFunc(flattenCmpFunc),
	}

	for _, testcase := range []struct {
		name  string
		sets  []mapping.DataInterval[blob]
		opts  []cfg.Option[mapping.Config[blob]]
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
		{
			name:  "Complex/MultipleOverlappingRanges/3Ranges",
			sets:  []mapping.DataInterval[blob]{iA, iB, iC},
			wants: i3RangesMerged,
		},
		{
			name:  "ComplexFlatten/MultipleOverlappingRanges/3Ranges",
			sets:  []mapping.DataInterval[blob]{iA, iB, iC},
			opts:  flattenOpts,
			wants: i3RangesFlattened,
		},
		{
			name:  "Complex/MultipleOverlappingRanges/4Ranges",
			sets:  []mapping.DataInterval[blob]{iA, iB, iC, iD},
			wants: i4RangesMerged,
		},
		{
			name:  "Complex/MultipleOverlappingRanges/5Ranges",
			sets:  []mapping.DataInterval[blob]{iA, iB, iC, iD, iE},
			wants: i5RangesMerged,
		},
		{
			name:  "Complex/MultipleOverlappingRanges/6Ranges",
			sets:  []mapping.DataInterval[blob]{iA, iB, iC, iD, iE, iF},
			wants: i6RangesMerged,
		},
		{
			name:  "ComplexFlatten/MultipleOverlappingRanges/6Ranges",
			sets:  []mapping.DataInterval[blob]{iA, iB, iC, iD, iE, iF},
			opts:  flattenOpts,
			wants: i6WantsFlattened,
		},
		{
			name:  "ComplexFlatten/MultipleStackedRanges/4Ranges",
			sets:  []mapping.DataInterval[blob]{iStackA, iStackB, iStackC, iStackD},
			opts:  flattenOpts,
			wants: iStackFlattened,
		},
	} {
		t.Run("Timeframe/"+testcase.name, func(t *testing.T) {
			tf := mapping.NewTimeframe[blob]()
			if !tf.AddSlice(testcase.sets...) {
				t.Fatalf("expected AddSlice to succeed")
			}

			organized := tf.Organize(testcase.opts...)
			assertIntervalsEqual(t, testcase.wants, organized.Slice())
			assertIntervalsEqual(t, testcase.wants, mapping.Collect(organized.All()))
		})

		t.Run("TimeframeStream/"+testcase.name, func(t *testing.T) {
			stream := mapping.NewTimeframeStream[blob](testcase.opts...)
			if !stream.Append(mapping.AsSeq(testcase.sets)) {
				t.Fatalf("expected Append to succeed")
			}

			stream = stream.Organize()
			assertIntervalsEqual(t, testcase.wants, stream.Slice())
			assertIntervalsEqual(t, testcase.wants, mapping.Collect(stream.All()))
		})

		t.Run("Organize/"+testcase.name, func(t *testing.T) {
			got := mapping.Organize(testcase.sets, testcase.opts...)
			assertIntervalsEqual(t, testcase.wants, got)
		})

		t.Run("OrganizeSeq/"+testcase.name, func(t *testing.T) {
			seq := mapping.OrganizeSeq(mapping.AsSeq(testcase.sets), testcase.opts...)
			assertIntervalsEqual(t, testcase.wants, mapping.Collect(seq))
		})
	}
}

func TestFlattenSections(t *testing.T) {
	type section struct {
		interval mapping.Interval
		data     string
	}

	type userSections struct {
		dataA []section
		dataB []section
		dataC []section
	}

	type sectionData struct {
		sectionA string
		sectionB string
		sectionC string
	}

	fn := func(u userSections) []mapping.DataInterval[sectionData] {
		userData := make([]mapping.DataInterval[sectionData], 0, len(u.dataA)+len(u.dataB)+len(u.dataC))

		for i := range u.dataA {
			userData = append(userData, mapping.DataInterval[sectionData]{
				Interval: u.dataA[i].interval,
				Data:     sectionData{sectionA: u.dataA[i].data},
			})
		}

		for i := range u.dataB {
			userData = append(userData, mapping.DataInterval[sectionData]{
				Interval: u.dataB[i].interval,
				Data:     sectionData{sectionB: u.dataB[i].data},
			})
		}

		for i := range u.dataC {
			userData = append(userData, mapping.DataInterval[sectionData]{
				Interval: u.dataC[i].interval,
				Data:     sectionData{sectionC: u.dataC[i].data},
			})
		}

		return userData
	}

	mergeFunc := func(a, b sectionData) sectionData {
		if b.sectionA != "" {
			a.sectionA = b.sectionA
		}
		if b.sectionB != "" {
			a.sectionB = b.sectionB
		}
		if b.sectionC != "" {
			a.sectionC = b.sectionC
		}

		return a
	}

	reducer := mapping.NewReducer(mapping.WithFlatten(mergeFunc))

	for _, testcase := range []struct {
		name  string
		input userSections
		wants []mapping.DataInterval[sectionData]
	}{
		{
			name:  "NoOp",
			input: userSections{},
			wants: []mapping.DataInterval[sectionData]{},
		},
		{
			name: "OneSection",
			input: userSections{
				dataA: []section{{
					interval: mapping.Interval{
						From: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
						To:   time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC),
					},
					data: "a",
				}},
			},
			wants: []mapping.DataInterval[sectionData]{{
				Interval: mapping.Interval{
					From: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC),
				},
				Data: sectionData{sectionA: "a"},
			}},
		},
		{
			name: "TwoSections/InSequence",
			input: userSections{
				dataA: []section{
					{
						interval: mapping.Interval{
							From: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
							To:   time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC),
						},
						data: "a",
					},
					{
						interval: mapping.Interval{
							From: time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC),
							To:   time.Date(2024, 1, 13, 0, 0, 0, 0, time.UTC),
						},
						data: "b",
					},
				},
			},
			wants: []mapping.DataInterval[sectionData]{
				{
					Interval: mapping.Interval{
						From: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
						To:   time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC),
					},
					Data: sectionData{sectionA: "a"},
				},
				{
					Interval: mapping.Interval{
						From: time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC),
						To:   time.Date(2024, 1, 13, 0, 0, 0, 0, time.UTC),
					},
					Data: sectionData{sectionA: "b"},
				},
			},
		},
		{
			name: "TwoSections/Overlapping",
			input: userSections{
				dataA: []section{{
					interval: mapping.Interval{
						From: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
						To:   time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC),
					},
					data: "a",
				}},
				dataB: []section{{
					interval: mapping.Interval{
						From: time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC),
						To:   time.Date(2024, 1, 13, 0, 0, 0, 0, time.UTC),
					},
					data: "b",
				}},
			},
			wants: []mapping.DataInterval[sectionData]{
				{
					Interval: mapping.Interval{
						From: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
						To:   time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC),
					},
					Data: sectionData{sectionA: "a"},
				},
				{
					Interval: mapping.Interval{
						From: time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC),
						To:   time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC),
					},
					Data: sectionData{sectionA: "a", sectionB: "b"},
				},
				{
					Interval: mapping.Interval{
						From: time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC),
						To:   time.Date(2024, 1, 13, 0, 0, 0, 0, time.UTC),
					},
					Data: sectionData{sectionB: "b"},
				},
			},
		},
		{
			name: "TwoSections/NextStartsBefore",
			input: userSections{
				dataA: []section{{
					interval: mapping.Interval{
						From: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
						To:   time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC),
					},
					data: "a",
				}},
				dataB: []section{{
					interval: mapping.Interval{
						From: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
						To:   time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
					},
					data: "b",
				}},
			},
			wants: []mapping.DataInterval[sectionData]{
				{
					Interval: mapping.Interval{
						From: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
						To:   time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
					},
					Data: sectionData{sectionB: "b"},
				},
				{
					Interval: mapping.Interval{
						From: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
						To:   time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC),
					},
					Data: sectionData{sectionA: "a"},
				},
			},
		},
		{
			name: "MultipleOverlappingSections",
			input: userSections{
				dataA: []section{{
					interval: mapping.Interval{
						From: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
						To:   time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC),
					},
					data: "a",
				}},
				dataB: []section{{
					interval: mapping.Interval{
						From: time.Date(2024, 1, 13, 0, 0, 0, 0, time.UTC),
						To:   time.Date(2024, 1, 17, 0, 0, 0, 0, time.UTC),
					},
					data: "b",
				}},
				dataC: []section{{
					interval: mapping.Interval{
						From: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
						To:   time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
					},
					data: "c",
				}},
			},
			wants: []mapping.DataInterval[sectionData]{
				{
					Interval: mapping.Interval{
						From: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
						To:   time.Date(2024, 1, 13, 0, 0, 0, 0, time.UTC),
					},
					Data: sectionData{sectionA: "a"},
				},
				{
					Interval: mapping.Interval{
						From: time.Date(2024, 1, 13, 0, 0, 0, 0, time.UTC),
						To:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
					},
					Data: sectionData{sectionA: "a", sectionB: "b"},
				},
				{
					Interval: mapping.Interval{
						From: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
						To:   time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
					},
					Data: sectionData{sectionA: "a", sectionB: "b", sectionC: "c"},
				},
				{
					Interval: mapping.Interval{
						From: time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
						To:   time.Date(2024, 1, 17, 0, 0, 0, 0, time.UTC),
					},
					Data: sectionData{sectionA: "a", sectionB: "b"},
				},
				{
					Interval: mapping.Interval{
						From: time.Date(2024, 1, 17, 0, 0, 0, 0, time.UTC),
						To:   time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC),
					},
					Data: sectionData{sectionA: "a"},
				},
			},
		},
		{
			name: "MultipleOverlappingSections/Complex",
			input: userSections{
				dataA: []section{
					{
						interval: mapping.Interval{
							From: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
							To:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
						},
						data: "a",
					},
					{
						interval: mapping.Interval{
							From: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
							To:   time.Date(2024, 1, 18, 0, 0, 0, 0, time.UTC),
						},
						data: "aa",
					},
					{
						interval: mapping.Interval{
							From: time.Date(2024, 1, 18, 0, 0, 0, 0, time.UTC),
							To:   time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC),
						},
						data: "aaa",
					},
				},
				dataB: []section{
					{
						interval: mapping.Interval{
							From: time.Date(2024, 1, 13, 0, 0, 0, 0, time.UTC),
							To:   time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC),
						},
						data: "b",
					},
					{
						interval: mapping.Interval{
							From: time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC),
							To:   time.Date(2024, 1, 17, 0, 0, 0, 0, time.UTC),
						},
						data: "bb",
					},
				},
				dataC: []section{
					{
						interval: mapping.Interval{
							From: time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC),
							To:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
						},
						data: "c",
					},
					{
						interval: mapping.Interval{
							From: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
							To:   time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
						},
						data: "cc",
					},
				},
			},
			wants: []mapping.DataInterval[sectionData]{
				{
					Interval: mapping.Interval{
						From: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
						To:   time.Date(2024, 1, 13, 0, 0, 0, 0, time.UTC),
					},
					Data: sectionData{sectionA: "a"},
				},
				{
					Interval: mapping.Interval{
						From: time.Date(2024, 1, 13, 0, 0, 0, 0, time.UTC),
						To:   time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC),
					},
					Data: sectionData{sectionA: "a", sectionB: "b"},
				},
				{
					Interval: mapping.Interval{
						From: time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC),
						To:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
					},
					Data: sectionData{sectionA: "a", sectionB: "bb", sectionC: "c"},
				},
				{
					Interval: mapping.Interval{
						From: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
						To:   time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
					},
					Data: sectionData{sectionA: "aa", sectionB: "bb", sectionC: "cc"},
				},
				{
					Interval: mapping.Interval{
						From: time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
						To:   time.Date(2024, 1, 17, 0, 0, 0, 0, time.UTC),
					},
					Data: sectionData{sectionA: "aa", sectionB: "bb"},
				},
				{
					Interval: mapping.Interval{
						From: time.Date(2024, 1, 17, 0, 0, 0, 0, time.UTC),
						To:   time.Date(2024, 1, 18, 0, 0, 0, 0, time.UTC),
					},
					Data: sectionData{sectionA: "aa"},
				},
				{
					Interval: mapping.Interval{
						From: time.Date(2024, 1, 18, 0, 0, 0, 0, time.UTC),
						To:   time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC),
					},
					Data: sectionData{sectionA: "aaa"},
				},
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			sections := fn(testcase.input)
			seq := reducer(mapping.AsSeq(sections))
			output := mapping.Collect(seq)

			assertIntervalsEqual(t, testcase.wants, output)
		})
	}
}

func TestTimeframeNonComparableMapValues(t *testing.T) {
	interval1 := mapping.Interval{
		From: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}
	interval2 := mapping.Interval{
		From: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
	}
	interval3 := mapping.Interval{
		From: time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
	}
	interval3Split1 := mapping.Interval{
		From: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC),
	}
	interval3Split2 := mapping.Interval{
		From: time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}
	interval6 := mapping.Interval{
		From: time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC),
	}
	interval7 := mapping.Interval{
		From: time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 22, 0, 0, 0, time.UTC),
	}
	interval8 := mapping.Interval{
		From: time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 21, 0, 0, 0, time.UTC),
	}

	kv1 := map[string]string{"a": "value", "b": "value"}
	kv2 := map[string]string{"c": "value", "d": "value"}
	kv3 := map[string]string{"c": "value"}
	kv4 := map[string]string{"d": "value"}
	kv5 := map[string]string{"e": "value"}

	mapMergeFunc := func(a, b map[string]string) map[string]string {
		out := make(map[string]string, len(a)+len(b))
		maps.Copy(out, a)
		maps.Copy(out, b)

		return out
	}

	opts := []cfg.Option[mapping.Config[map[string]string]]{
		mapping.WithMergeFunc(mapMergeFunc),
		mapping.WithCompareFunc(maps.Equal[map[string]string, map[string]string]),
	}

	for _, testcase := range []struct {
		name  string
		input []mapping.DataInterval[map[string]string]
		wants []mapping.DataInterval[map[string]string]
	}{
		{
			name: "Sequential",
			input: []mapping.DataInterval[map[string]string]{
				{Interval: interval1, Data: kv1},
				{Interval: interval2, Data: kv2},
			},
			wants: []mapping.DataInterval[map[string]string]{
				{Interval: interval1, Data: kv1},
				{Interval: interval2, Data: kv2},
			},
		},
		{
			name: "InterleavedOnTail",
			input: []mapping.DataInterval[map[string]string]{
				{Interval: interval1, Data: kv1},
				{Interval: interval3, Data: kv2},
			},
			wants: []mapping.DataInterval[map[string]string]{
				{Interval: interval3Split1, Data: kv1},
				{Interval: interval3Split2, Data: map[string]string{"a": "value", "b": "value", "c": "value", "d": "value"}},
				{Interval: interval2, Data: kv2},
			},
		},
		{
			name: "InterleavedRecursive",
			input: []mapping.DataInterval[map[string]string]{
				{Interval: interval1, Data: kv1},
				{Interval: interval6, Data: kv3},
				{Interval: interval7, Data: kv4},
				{Interval: interval8, Data: kv5},
			},
			wants: []mapping.DataInterval[map[string]string]{
				{Interval: mapping.Interval{From: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC)}, Data: kv1},
				{Interval: mapping.Interval{From: time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC)}, Data: map[string]string{"a": "value", "b": "value", "c": "value"}},
				{Interval: mapping.Interval{From: time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC)}, Data: map[string]string{"a": "value", "b": "value", "c": "value", "d": "value"}},
				{Interval: interval8, Data: map[string]string{"a": "value", "b": "value", "c": "value", "d": "value", "e": "value"}},
				{Interval: mapping.Interval{From: time.Date(2024, 1, 1, 21, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 1, 22, 0, 0, 0, time.UTC)}, Data: map[string]string{"a": "value", "b": "value", "c": "value", "d": "value"}},
				{Interval: mapping.Interval{From: time.Date(2024, 1, 1, 22, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC)}, Data: map[string]string{"a": "value", "b": "value", "c": "value"}},
				{Interval: mapping.Interval{From: time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC), To: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)}, Data: kv1},
			},
		},
	} {
		t.Run("Timeframe/"+testcase.name, func(t *testing.T) {
			tf := mapping.NewTimeframe[map[string]string](opts...)
			tf.AddSlice(testcase.input...)
			got := tf.Organize().Slice()

			if len(got) != len(testcase.wants) {
				t.Fatalf("length mismatch: want %d, got %d", len(testcase.wants), len(got))
			}

			for i := range testcase.wants {
				if got[i].Interval != testcase.wants[i].Interval {
					t.Fatalf("[%d] interval mismatch: want %v, got %v", i, testcase.wants[i].Interval, got[i].Interval)
				}
				if !maps.Equal(got[i].Data, testcase.wants[i].Data) {
					t.Fatalf("[%d] data mismatch: want %v, got %v", i, testcase.wants[i].Data, got[i].Data)
				}
			}
		})

		t.Run("TimeframeStream/"+testcase.name, func(t *testing.T) {
			stream := mapping.NewTimeframeStream[map[string]string](opts...)
			stream.AddSlice(testcase.input...)
			got := stream.Slice()

			if len(got) != len(testcase.wants) {
				t.Fatalf("length mismatch: want %d, got %d", len(testcase.wants), len(got))
			}

			for i := range testcase.wants {
				if got[i].Interval != testcase.wants[i].Interval {
					t.Fatalf("[%d] interval mismatch: want %v, got %v", i, testcase.wants[i].Interval, got[i].Interval)
				}
				if !maps.Equal(got[i].Data, testcase.wants[i].Data) {
					t.Fatalf("[%d] data mismatch: want %v, got %v", i, testcase.wants[i].Data, got[i].Data)
				}
			}
		})
	}
}

func TestTimeframeEdgeCasesAndMerge(t *testing.T) {
	var nilTF *mapping.Timeframe[int]
	if nilTF.Add(mapping.Interval{}, 1) {
		t.Fatalf("expected nil Timeframe.Add to return false")
	}
	if nilTF.AddSlice() {
		t.Fatalf("expected nil Timeframe.AddSlice to return false")
	}
	if nilTF.Append(nil) {
		t.Fatalf("expected nil Timeframe.Append to return false")
	}
	if nilTF.Len() != 0 {
		t.Fatalf("expected nil Timeframe.Len() == 0")
	}
	if len(nilTF.Slice()) != 0 {
		t.Fatalf("expected nil Timeframe.Slice() to be empty")
	}
	for range nilTF.All() {
		t.Fatalf("expected nil Timeframe.All() to yield 0 items")
	}
	if nilTF.Organize().Len() != 0 {
		t.Fatalf("expected nil Timeframe.Organize() to return empty Timeframe")
	}
	if nilTF.Merge(nil).Len() != 0 {
		t.Fatalf("expected nilTF.Merge(nil) to return empty Timeframe")
	}

	var nilStream *mapping.TimeframeStream[int]
	if nilStream.Add(mapping.Interval{}, 1) {
		t.Fatalf("expected nil TimeframeStream.Add to return false")
	}
	if nilStream.AddSlice() {
		t.Fatalf("expected nil TimeframeStream.AddSlice to return false")
	}
	if nilStream.Append(nil) {
		t.Fatalf("expected nil TimeframeStream.Append to return false")
	}
	if nilStream.Len() != 0 {
		t.Fatalf("expected nil TimeframeStream.Len() == 0")
	}
	if len(nilStream.Slice()) != 0 {
		t.Fatalf("expected nil TimeframeStream.Slice() to be empty")
	}
	for range nilStream.All() {
		t.Fatalf("expected nil TimeframeStream.All() to yield 0 items")
	}
	for range mapping.OrganizeSeq[int](nil) {
		t.Fatalf("expected OrganizeSeq(nil) to yield 0 items")
	}

	// Zero-value initialization and invalid intervals.
	var zeroTF mapping.Timeframe[int]
	var zeroStream mapping.TimeframeStream[int]

	invalid := mapping.Interval{
		From: time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
	}
	zeroWidth := mapping.Interval{
		From: time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
	}
	valid1 := mapping.Interval{
		From: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
	}
	valid2 := mapping.Interval{
		From: time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 10, 16, 0, 0, 0, time.UTC),
	}

	if zeroTF.Add(invalid, 1) {
		t.Fatalf("expected invalid interval to be rejected by Timeframe.Add")
	}
	if zeroStream.Add(invalid, 1) {
		t.Fatalf("expected invalid interval to be rejected by TimeframeStream.Add")
	}
	if zeroStream.Add(zeroWidth, 1) {
		t.Fatalf("expected zero-width interval with offset 0 to be rejected by TimeframeStream.Add")
	}

	if zeroTF.AddSlice(mapping.DataInterval[int]{Interval: invalid, Data: 1}) {
		t.Fatalf("expected AddSlice with invalid interval to return false")
	}
	if zeroStream.AddSlice(mapping.DataInterval[int]{Interval: invalid, Data: 1}) {
		t.Fatalf("expected stream AddSlice with invalid interval to return false")
	}
	if zeroTF.Append(mapping.AsSeq([]mapping.DataInterval[int]{{Interval: invalid, Data: 1}})) {
		t.Fatalf("expected Append with invalid interval to return false")
	}
	if zeroStream.Append(mapping.AsSeq([]mapping.DataInterval[int]{{Interval: invalid, Data: 1}})) {
		t.Fatalf("expected stream Append with invalid interval to return false")
	}

	// Verify Organize and OrganizeSeq skip invalid intervals across strategies.
	mixedInput := []mapping.DataInterval[int]{
		{Interval: invalid, Data: 99},
		{Interval: zeroWidth, Data: 98},
		{Interval: valid1, Data: 10},
	}
	if got := mapping.Organize(mixedInput, mapping.WithFlatten[int](nil)); len(got) != 1 {
		t.Fatalf("expected 1 valid interval from Organize, got %d", len(got))
	}
	if got := mapping.Organize(mixedInput, mapping.WithAppend[int]()); len(got) != 1 {
		t.Fatalf("expected 1 valid interval from Organize(WithAppend), got %d", len(got))
	}
	if got := mapping.Collect(mapping.OrganizeSeq(mapping.AsSeq(mixedInput), mapping.WithReplace[int]())); len(got) != 1 {
		t.Fatalf("expected 1 valid interval from OrganizeSeq, got %d", len(got))
	}
	if got := mapping.Collect(mapping.OrganizeSeq(mapping.AsSeq(mixedInput), mapping.WithAppend[int]())); len(got) != 1 {
		t.Fatalf("expected 1 valid interval from OrganizeSeq(WithAppend), got %d", len(got))
	}

	if !zeroTF.Append(mapping.AsSeq([]mapping.DataInterval[int]{{Interval: valid1, Data: 10}})) {
		t.Fatalf("expected valid interval to be appended to zero Timeframe")
	}
	if !zeroStream.Add(valid1, 10) {
		t.Fatalf("expected valid interval to be added to zero TimeframeStream")
	}
	if zeroStream.Len() != 1 {
		t.Fatalf("expected stream Len() == 1, got %d", zeroStream.Len())
	}

	if got := nilTF.Merge(&zeroTF); got.Len() != 1 {
		t.Fatalf("expected nilTF.Merge(&zeroTF) to have Len 1, got %d", got.Len())
	}
	if got := zeroTF.Merge(nil); got.Len() != 1 {
		t.Fatalf("expected zeroTF.Merge(nil) to have Len 1, got %d", got.Len())
	}

	tf2 := mapping.NewTimeframe[int]()
	tf2.Add(valid2, 20)

	merged := zeroTF.Merge(tf2, mapping.WithFlatten(func(a, b int) int { return a + b }))
	wants := []mapping.DataInterval[int]{
		{
			Interval: mapping.Interval{
				From: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
				To:   time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
			},
			Data: 10,
		},
		{
			Interval: mapping.Interval{
				From: time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
				To:   time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
			},
			Data: 30,
		},
		{
			Interval: mapping.Interval{
				From: time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC),
				To:   time.Date(2024, 1, 10, 16, 0, 0, 0, time.UTC),
			},
			Data: 20,
		},
	}

	assertIntervalsEqual(t, wants, merged.Slice())
}

func assertIntervalsEqual[T comparable](t *testing.T, wants, got []mapping.DataInterval[T]) {
	t.Helper()

	if !slices.Equal(wants, got) {
		t.Fatalf("intervals mismatch:\nwant: %+v\ngot:  %+v", wants, got)
	}
}
