package mapping

import (
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

	kv1 := map[string]string{
		"a": "value",
		"b": "value",
	}

	kv2 := map[string]string{
		"c": "value",
		"d": "value",
	}

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
	} {
		t.Run(testcase.name, func(t *testing.T) {
			tf := NewTimeframe[string, string]()

			for interval, values := range testcase.input {
				err := tf.Add(interval, values)
				isEqual(t, nil, err)
			}

			for i := range testcase.print {
				itf, ok := tf.Index.values[testcase.print[i]]
				isEqual(t, true, ok)

				t.Log(itf)
			}
		})
	}
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
