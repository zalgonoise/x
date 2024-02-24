package mapping

import (
	"testing"
	"time"
)

func TestTimeframe(t *testing.T) {
	interval1 := Interval{
		From: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Dur:  12 * time.Hour,
	}
	interval2 := Interval{
		From: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		Dur:  12 * time.Hour,
	}

	kv1 := []KV[string, string]{
		{true, "a", "value"},
		{true, "b", "value"},
	}

	kv2 := []KV[string, string]{
		{true, "c", "value"},
		{true, "d", "value"},
	}

	for _, testcase := range []struct {
		name  string
		input []KV[Interval, []KV[string, string]]
		print []Interval
	}{
		{
			name: "sequential",
			input: []KV[Interval, []KV[string, string]]{
				{true, interval1, kv1},
				{true, interval2, kv2},
			},
			print: []Interval{interval1, interval2},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			tf := NewTimeframe[string, string]()

			for i := range testcase.input {
				err := tf.Add(testcase.input[i].Key, testcase.input[i].Value)
				isEqual(t, nil, err)
			}

			newTF, err := tf.Organize(tf.All())
			isEqual(t, nil, err)

			for i := range testcase.print {
				itf, ok := newTF.Index.values[testcase.print[i]]
				isEqual(t, true, ok)

				t.Log(itf.Keys)
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
