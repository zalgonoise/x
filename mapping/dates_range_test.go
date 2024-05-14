package mapping_test

import (
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zalgonoise/x/mapping"
)

func TestTimeframeRange(t *testing.T) {
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
			next.user.id = cur.user.id
			next.data.len = cur.data.len
		}

		return next
	}

	flattenCmpFunc := func(cur, next blob) bool {
		return cur.user.name == next.user.name
	}

	bA := blob{
		user: user{id: 1, name: "blob-a"},
		data: data{len: 1},
	}

	bB := blob{
		user: user{id: 2, name: "blob-b"},
		data: data{len: 1},
	}

	bC := blob{
		user: user{id: 3, name: "blob-c"},
		data: data{len: 1},
	}

	bD := blob{
		user: user{id: 4, name: "blob-d"},
		data: data{len: 1},
	}

	bE := blob{
		user: user{id: 5, name: "blob-e"},
		data: data{len: 1},
	}

	bF := blob{
		user: user{id: 6, name: "blob-f"},
		data: data{len: 1},
	}

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

	// complex test - 3 ranges
	//
	// |###### A ######|
	//    |#### B ####################|
	//         |# C #|

	i3RangesMerged1 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 5, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 6, 0, 0, 0, time.UTC),
		},
		Data: bA,
	}

	i3RangesMerged2 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 6, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 7, 0, 0, 0, time.UTC),
		},
		Data: bB,
	}

	i3RangesMerged3 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 7, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC),
		},
		Data: bC,
	}

	i3RangesMerged4 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC),
		},
		Data: bB,
	}

	// complex test - 3 ranges (flattened)
	//
	// |###### A #######|
	//    |#### B ####################|
	//         |# C #|
	//
	// |A | -B | --C |-B|      B      |

	i3RangesFlattened1 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 5, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 6, 0, 0, 0, time.UTC),
		},
		Data: bA,
	}

	i3RangesFlattened2 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 6, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 7, 0, 0, 0, time.UTC),
		},
		Data: blob{
			user: user{id: 1, name: "blob-b"},
			data: data{len: 1},
		},
	}

	i3RangesFlattened3 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 7, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC),
		},
		Data: blob{
			user: user{id: 1, name: "blob-c"},
			data: data{len: 1},
		},
	}

	i3RangesFlattened4 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC),
		},
		Data: blob{
			user: user{id: 1, name: "blob-b"},
			data: data{len: 1},
		},
	}

	// complex test - 4 ranges
	//
	// |###### A ######|
	//    |#### B ####################|
	//         |# C #|
	//                    |# D #|

	i4RangesMerged1 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 5, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 6, 0, 0, 0, time.UTC),
		},
		Data: bA,
	}

	i4RangesMerged2 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 6, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 7, 0, 0, 0, time.UTC),
		},
		Data: bB,
	}

	i4RangesMerged3 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 7, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC),
		},
		Data: bC,
	}

	i4RangesMerged4 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
		},
		Data: bB,
	}

	i4RangesMerged5 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
		},
		Data: bD,
	}

	i4RangesMerged6 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC),
		},
		Data: bB,
	}

	// complex test - 5 ranges
	//
	// |###### A ######|
	//    |#### B ####################|
	//         |# C #|
	//                    |# D #|
	//                        |######### E ##########|

	i5RangesMerged1 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 5, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 6, 0, 0, 0, time.UTC),
		},
		Data: bA,
	}

	i5RangesMerged2 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 6, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 7, 0, 0, 0, time.UTC),
		},
		Data: bB,
	}

	i5RangesMerged3 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 7, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC),
		},
		Data: bC,
	}

	i5RangesMerged4 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
		},
		Data: bB,
	}

	i5RangesMerged5 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 11, 0, 0, 0, time.UTC),
		},
		Data: bD,
	}

	i5RangesMerged6 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 11, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 20, 0, 0, 0, time.UTC),
		},
		Data: bE,
	}

	// complex test - 6 ranges
	//
	// |###### A ######|
	//    |#### B ####################|
	//         |# C #|
	//                    |# D #|
	//                        |######### E ##########|
	//                                      |# F #|

	i6RangesMerged1 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 5, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 6, 0, 0, 0, time.UTC),
		},
		Data: bA,
	}

	i6RangesMerged2 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 6, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 7, 0, 0, 0, time.UTC),
		},
		Data: bB,
	}

	i6RangesMerged3 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 7, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC),
		},
		Data: bC,
	}

	i6RangesMerged4 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
		},
		Data: bB,
	}

	i6RangesMerged5 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 11, 0, 0, 0, time.UTC),
		},
		Data: bD,
	}

	i6RangesMerged6 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 11, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 17, 0, 0, 0, time.UTC),
		},
		Data: bE,
	}

	i6RangesMerged7 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 17, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 18, 0, 0, 0, time.UTC),
		},
		Data: bF,
	}

	i6RangesMerged8 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 18, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 20, 0, 0, 0, time.UTC),
		},
		Data: bE,
	}

	// complex test - 6 ranges (flattened)
	//
	// |###### A #######|
	//    |#### B ####################|
	//         |# C #|
	//                    |## D ##|
	//                        |######### E ##########|
	//                                      |# F #|
	//
	// |A | -B | --C |-B|B| -D|--E| -E|  E  |  -F | E|

	i6RangesFlattened1 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 5, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 6, 0, 0, 0, time.UTC),
		},
		Data: bA,
	}

	i6RangesFlattened2 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 6, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 7, 0, 0, 0, time.UTC),
		},
		Data: blob{
			user: user{id: 1, name: "blob-b"},
			data: data{len: 1},
		},
	}

	i6RangesFlattened3 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 7, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC),
		},
		Data: blob{
			user: user{id: 1, name: "blob-c"},
			data: data{len: 1},
		},
	}

	i6RangesFlattened4 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
		},
		Data: blob{
			user: user{id: 1, name: "blob-b"},
			data: data{len: 1},
		},
	}

	i6RangesFlattened5 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 11, 0, 0, 0, time.UTC),
		},
		Data: blob{
			user: user{id: 1, name: "blob-d"},
			data: data{len: 1},
		},
	}

	i6RangesFlattened6 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 11, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 17, 0, 0, 0, time.UTC),
		},
		Data: bE,
	}

	i6RangesFlattened7 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 17, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 18, 0, 0, 0, time.UTC),
		},
		Data: blob{
			user: user{id: 5, name: "blob-f"},
			data: data{len: 1},
		},
	}

	i6RangesFlattened8 := mapping.DataInterval[blob]{
		Interval: mapping.Interval{
			From: time.Date(2024, 1, 10, 18, 0, 0, 0, time.UTC),
			To:   time.Date(2024, 1, 10, 20, 0, 0, 0, time.UTC),
		},
		Data: bE,
	}

	for _, testcase := range []struct {
		name    string
		sets    []mapping.DataInterval[blob]
		reducer mapping.ReducerFunc[blob]
		wants   []mapping.DataInterval[blob]
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
		{
			name: "Complex/MultipleOverlappingRanges/3Ranges",
			sets: []mapping.DataInterval[blob]{
				iA, iB, iC,
			},
			wants: []mapping.DataInterval[blob]{
				i3RangesMerged1, i3RangesMerged2, i3RangesMerged3, i3RangesMerged4,
			},
		},
		{
			name: "ComplexFlatten/MultipleOverlappingRanges/3Ranges",
			sets: []mapping.DataInterval[blob]{
				iA, iB, iC,
			},
			reducer: mapping.Flatten(flattenCmpFunc, flattenMergeFunc),
			wants: []mapping.DataInterval[blob]{
				i3RangesFlattened1, i3RangesFlattened2, i3RangesFlattened3, i3RangesFlattened4,
			},
		},
		{
			name: "Complex/MultipleOverlappingRanges/4Ranges",
			sets: []mapping.DataInterval[blob]{
				iA, iB, iC, iD,
			},
			wants: []mapping.DataInterval[blob]{
				i4RangesMerged1, i4RangesMerged2, i4RangesMerged3, i4RangesMerged4, i4RangesMerged5, i4RangesMerged6,
			},
		},
		{
			name: "Complex/MultipleOverlappingRanges/5Ranges",
			sets: []mapping.DataInterval[blob]{
				iA, iB, iC, iD, iE,
			},
			wants: []mapping.DataInterval[blob]{
				i5RangesMerged1, i5RangesMerged2, i5RangesMerged3, i5RangesMerged4, i5RangesMerged5, i5RangesMerged6,
			},
		},
		{
			name: "Complex/MultipleOverlappingRanges/6Ranges",
			sets: []mapping.DataInterval[blob]{
				iA, iB, iC, iD, iE, iF,
			},
			wants: []mapping.DataInterval[blob]{
				i6RangesMerged1, i6RangesMerged2, i6RangesMerged3, i6RangesMerged4, i6RangesMerged5, i6RangesMerged6,
				i6RangesMerged7, i6RangesMerged8,
			},
		},
		{
			name: "ComplexFlatten/MultipleOverlappingRanges/6Ranges",
			sets: []mapping.DataInterval[blob]{
				iA, iB, iC, iD, iE, iF,
			},
			reducer: mapping.Flatten(flattenCmpFunc, flattenMergeFunc),
			wants: []mapping.DataInterval[blob]{
				i6RangesFlattened1, i6RangesFlattened2, i6RangesFlattened3, i6RangesFlattened4, i6RangesFlattened5,
				i6RangesFlattened6, i6RangesFlattened7, i6RangesFlattened8,
			},
		},
	} {
		t.Run("InitTimeframeRange/"+testcase.name, func(t *testing.T) {
			tf := mapping.NewTimeframeRange[blob]()

			for i := range testcase.sets {
				_ = tf.Add(testcase.sets[i].Interval, testcase.sets[i].Data)
			}

			if testcase.reducer == nil {
				testcase.reducer = mapping.Replace[blob]()
			}

			newTF := tf.Organize(testcase.reducer)

			seq := newTF.All()

			require.True(t, seq(verifySeq(t, testcase.wants)))
		})

		t.Run("OrganizeTimeframeRange/"+testcase.name, func(t *testing.T) {
			if testcase.reducer == nil {
				testcase.reducer = mapping.Replace[blob]()
			}

			tf := mapping.Organize[*mapping.TimeframeRange[blob]](mapping.AsSeq(testcase.sets), testcase.reducer)

			require.True(t, tf.All()(verifySeq(t, testcase.wants)))
		})
	}
}

func verifySeq(t *testing.T, wants []mapping.DataInterval[blob]) func(interval mapping.Interval, value blob) bool {
	return func(interval mapping.Interval, value blob) bool {
		var zero blob

		if value == zero {
			return false
		}

		idx := slices.IndexFunc(wants, func(set mapping.DataInterval[blob]) bool {
			return set.Interval == interval
		})

		if idx < 0 {
			t.Error("interval not present in expected results", interval)

			return false
		}

		if value != wants[idx].Data {
			t.Error("value doesn't match expected", value, wants[idx].Data)

			return false
		}

		return true
	}
}
