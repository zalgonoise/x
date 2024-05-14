package mapping

import (
	"maps"
	"slices"
)

const minAlloc = 64

// TimeframeMap stores values in intervals of time, as a 2D-map of Interval to a map of types K and T.
type TimeframeMap[K comparable, T any] struct {
	Keys   []Interval
	Values map[Interval]map[K]T
}

// NewTimeframeMap creates a TimeframeMap of types K and T, with an 2D-map of Interval's to a map of types K and T.
func NewTimeframeMap[K comparable, T any]() *TimeframeMap[K, T] {
	return &TimeframeMap[K, T]{
		Keys:   make([]Interval, 0, minAlloc),
		Values: make(map[Interval]map[K]T, minAlloc),
	}
}

func (t *TimeframeMap[K, T]) init() *TimeframeMap[K, T] {
	if t == nil {
		return NewTimeframeMap[K, T]()
	}

	if t.Keys == nil {
		t.Keys = make([]Interval, 0, minAlloc)
	}

	if t.Values == nil {
		t.Values = make(map[Interval]map[K]T, minAlloc)
	}

	return t
}

// Add joins the Interval i and its values to the TimeframeMap t, while ordering its
// previously inserted Interval(s) in the process.
func (t *TimeframeMap[K, T]) Add(i Interval, values map[K]T) bool {
	if val, ok := t.Values[i]; ok {
		t.Values[i] = coalesce(val, values)

		return false
	}

	t.Keys = append(t.Keys, i)
	t.Values[i] = values

	return true
}

// Append iterates through the input SeqKV and adds all intervals and respective values
// to the TimeframeMap t.
func (t *TimeframeMap[K, T]) Append(seq SeqKV[Interval, map[K]T]) error {
	if !seq(t.Add) {
		return ErrAppendFailed
	}

	return nil
}

// All returns an iterator over the values in the TimeframeMap,
// through the indexed Interval values ordered by their From time.Time value.
func (t *TimeframeMap[K, T]) All() SeqKV[Interval, map[K]T] {
	return func(yield func(Interval, map[K]T) bool) bool {
		slices.SortFunc(t.Keys, func(a, b Interval) int {
			return a.From.Compare(b.From)
		})

		for i := range t.Keys {
			values, ok := t.Values[t.Keys[i]]
			if !ok {
				continue
			}

			if !yield(t.Keys[i], values) {
				return false
			}
		}

		return true
	}
}

// Organize returns a new TimeframeMap with organized Interval(s) and respective values. It is the result of
// calling Flatten on TimeframeMap.All, and appending the resulting sequence to a new instance of TimeframeMap.
func (t *TimeframeMap[K, T]) Organize(cmp func(a, b T) bool) *TimeframeMap[K, T] {
	seq := Flatten(cmpFunc[K](cmp), mergeFunc[K, T])(t.All())

	tf := NewTimeframeMap[K, T]()

	seq(func(interval Interval, t map[K]T) bool {
		_ = tf.Add(interval, t)

		return true
	})

	slices.SortFunc(tf.Keys, func(a, b Interval) int {
		return a.From.Compare(b.From)
	})

	return tf
}

// Merge joins the intervals and respective values of the TimeframeMap tf into the TimeframeMap t,
// by extracting a SeqKV of the same items from tf using TimeframeMap.All, and adding them into
// Timeframe t using TimeframeMap.Append.
func (t *TimeframeMap[K, T]) Merge(tf *Timeframe[K, T]) (err error) {
	return t.Append(tf.All())
}

func mergeFunc[K comparable, T any](a, b map[K]T) map[K]T {
	dataCopy := maps.Clone(a)
	coalesce(dataCopy, b)

	return dataCopy
}

func cmpFunc[K comparable, T any](cmp func(a, b T) bool) func(a, b map[K]T) bool {
	return func(a, b map[K]T) bool {
		if len(a) != len(b) {
			return false
		}

		for k, v1 := range a {
			if v2, ok := b[k]; !ok || !cmp(v1, v2) {
				return false
			}
		}
		return true
	}
}
