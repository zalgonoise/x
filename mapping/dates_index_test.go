package mapping_test

import (
	"maps"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zalgonoise/x/mapping"
)

func TestTimeframeIndex(t *testing.T) {
	interval1 := mapping.Interval{
		From: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}
	interval2 := mapping.Interval{
		From: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
	}

	// interleaved on tail
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

	// interleaved on head
	interval4 := mapping.Interval{
		From: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC),
	}

	interval4Split1 := mapping.Interval{
		From: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	interval4Split2 := mapping.Interval{
		From: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC),
	}
	interval4Split3 := mapping.Interval{
		From: time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	// interleaved in the middle
	interval5 := mapping.Interval{
		From: time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 16, 0, 0, 0, time.UTC),
	}

	interval5Split1 := mapping.Interval{
		From: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
	}
	interval5Split2 := mapping.Interval{
		From: time.Date(2024, 1, 1, 16, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	// interleaved recursively
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

	interval6Split1 := mapping.Interval{
		From: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
	}
	interval6Split2 := mapping.Interval{
		From: time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC),
	}
	interval6Split3 := mapping.Interval{
		From: time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
	}

	interval6Split4 := mapping.Interval{
		From: time.Date(2024, 1, 1, 21, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 22, 0, 0, 0, time.UTC),
	}
	interval6Split5 := mapping.Interval{
		From: time.Date(2024, 1, 1, 22, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC),
	}
	interval6Split6 := mapping.Interval{
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

	mapMergeFunc := func(a, b map[string]string) map[string]string {
		aCopy := make(map[string]string, len(a))
		maps.Copy(aCopy, a)

		for k, v := range b {
			aCopy[k] = v
		}

		return aCopy
	}

	for _, testcase := range []struct {
		name  string
		input []mapping.DataInterval[map[string]string]
		wants []mapping.DataInterval[map[string]string]
	}{
		{
			name: "sequential",
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
			name: "interleaved/on_tail",
			input: []mapping.DataInterval[map[string]string]{
				{Interval: interval1, Data: kv1},
				{Interval: interval3, Data: kv2},
			},
			wants: []mapping.DataInterval[map[string]string]{
				{Interval: interval3Split1, Data: kv1},
				{Interval: interval3Split2, Data: map[string]string{
					"a": "value",
					"b": "value",
					"c": "value",
					"d": "value",
				}},
				{Interval: interval2, Data: kv2},
			},
		},
		{
			name: "interleaved/on_head",
			input: []mapping.DataInterval[map[string]string]{
				{Interval: interval1, Data: kv1},
				{Interval: interval4, Data: kv2},
			},
			wants: []mapping.DataInterval[map[string]string]{
				{Interval: interval4Split1, Data: kv2},
				{Interval: interval4Split2, Data: map[string]string{
					"a": "value",
					"b": "value",
					"c": "value",
					"d": "value",
				}},
				{Interval: interval4Split3, Data: kv1},
			},
		},
		{
			name: "interleaved/on_middle",
			input: []mapping.DataInterval[map[string]string]{
				{Interval: interval1, Data: kv1},
				{Interval: interval5, Data: kv2},
			},
			wants: []mapping.DataInterval[map[string]string]{
				{Interval: interval5Split1, Data: kv1},
				{Interval: interval5, Data: map[string]string{
					"a": "value",
					"b": "value",
					"c": "value",
					"d": "value",
				}},
				{Interval: interval5Split2, Data: kv1},
			},
		},
		{
			name: "interleaved/recursive",
			input: []mapping.DataInterval[map[string]string]{
				{Interval: interval1, Data: kv1},
				{Interval: interval6, Data: kv3},
				{Interval: interval7, Data: kv4},
				{Interval: interval8, Data: kv5},
			},
			wants: []mapping.DataInterval[map[string]string]{
				{Interval: interval6Split1, Data: kv1},
				{Interval: interval6Split2, Data: map[string]string{
					"a": "value",
					"b": "value",
					"c": "value",
				}},
				{Interval: interval6Split3, Data: map[string]string{
					"a": "value",
					"b": "value",
					"c": "value",
					"d": "value",
				}},
				{Interval: interval8, Data: map[string]string{
					"a": "value",
					"b": "value",
					"c": "value",
					"d": "value",
					"e": "value",
				}},
				{Interval: interval6Split4, Data: map[string]string{
					"a": "value",
					"b": "value",
					"c": "value",
					"d": "value",
				}},
				{Interval: interval6Split5, Data: map[string]string{
					"a": "value",
					"b": "value",
					"c": "value",
				}},
				{Interval: interval6Split6, Data: map[string]string{
					"a": "value",
					"b": "value",
				}},
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			tf := mapping.NewTimeframeIndex[map[string]string]()

			for i := range testcase.input {
				_ = tf.Add(testcase.input[i].Interval, testcase.input[i].Data)
			}

			tf = tf.Organize(mapping.FlattenFunc[map[string]string](
				maps.Equal[map[string]string, map[string]string],
				mapMergeFunc, 0))

			require.True(t, tf.All()(verifySeqMap(testcase.wants)))
		})

		t.Run("OrganizeTimeframeIndex/"+testcase.name, func(t *testing.T) {
			fn := func(yield func(mapping.Interval, map[string]string) bool) bool {
				for i := range testcase.input {
					if !yield(testcase.input[i].Interval, testcase.input[i].Data) {
						return false
					}
				}

				return true
			}

			tf := mapping.OrganizeSeq[*mapping.TimeframeIndex[map[string]string]](fn,
				mapping.FlattenFunc[map[string]string](
					maps.Equal[map[string]string, map[string]string],
					mapMergeFunc, 0))

			require.True(t, tf.All()(verifySeqMap(testcase.wants)))
		})
	}
}

func verifySeqMap(wants []mapping.DataInterval[map[string]string]) func(interval mapping.Interval, m map[string]string) bool {
	return func(interval mapping.Interval, b map[string]string) bool {

		idx := slices.IndexFunc(wants, func(set mapping.DataInterval[map[string]string]) bool {
			return set.Interval == interval
		})

		if idx < 0 {
			return false
		}

		if len(wants[idx].Data) != len(b) {
			return false
		}

		for k, v := range wants[idx].Data {
			if value, ok := b[k]; !ok || v != value {
				return false
			}
		}

		return true
	}
}
