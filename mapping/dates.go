package mapping

import (
	"errors"
	"time"
)

var (
	ErrAppendFailed    = errors.New("failed to append to timeframe")
	ErrTimeSplitFailed = errors.New("failed to split time intervals")
)

// Timeframe stores values in intervals of time, as an Index of Interval and a map of types K and T.
type Timeframe[K comparable, T any] struct {
	Index *Index[Interval, map[K]T]
}

// Interval is a period of time with a From and To time.Time values.
type Interval struct {
	From time.Time
	To   time.Time
}

// NewTimeframe creates a Timeframe of types K and T, with an index on the Interval's From time.Time.
func NewTimeframe[K comparable, T any]() *Timeframe[K, T] {
	return &Timeframe[K, T]{
		Index: NewIndex[Interval, map[K]T](
			make(map[Interval]map[K]T),
			WithIndex[map[K]T](func(a, b Interval) int {
				return a.From.Compare(b.From)
			}),
			WithZero[Interval, map[K]T](nil),
		),
	}
}

func (t *Timeframe[K, T]) init() *Timeframe[K, T] {
	if t == nil {
		return NewTimeframe[K, T]()
	}

	if t.Index == nil {
		t.Index = NewIndex[Interval, map[K]T](
			make(map[Interval]map[K]T),
			WithIndex[map[K]T](func(a, b Interval) int {
				return a.From.Compare(b.From)
			}),
			WithZero[Interval, map[K]T](nil),
		)
	}

	return t
}

// Add joins the Interval i and its values to the Timeframe t, while ordering its
// previously inserted Interval(s) in the process.
func (t *Timeframe[K, T]) Add(i Interval, values map[K]T) bool {
	if i.To.Before(i.From) {
		return false
	}

	if len(t.Index.Keys) == 0 {
		t.Index.Set(i, func(old map[K]T) (map[K]T, bool) {
			return values, true
		})

		return true
	}

	if _, ok := t.Index.Get(i); ok {
		t.Index.Set(i, func(old map[K]T) (map[K]T, bool) {
			return coalesce(old, values), true
		})

		return true
	}

	t.Index.Set(i, func(old map[K]T) (map[K]T, bool) {
		return values, true
	})

	return true
}

// Append iterates through the input SeqKV and adds all intervals and respective values
// to the Timeframe t.
func (t *Timeframe[K, T]) Append(seq SeqKV[Interval, map[K]T]) (err error) {
	if !seq(t.Add) {
		return ErrAppendFailed
	}

	return nil
}

// All returns an iterator over the values in the Timeframe,
// through the indexed Interval values ordered by their From time.Time value.
func (t *Timeframe[K, T]) All() SeqKV[Interval, map[K]T] {
	return func(yield func(Interval, map[K]T) bool) bool {
		keys := t.Index.Keys

		for i := range keys {
			values, ok := t.Index.values[keys[i]]
			if !ok {
				continue
			}

			if !yield(keys[i], values) {
				return false
			}
		}

		return true
	}
}

// Organize returns a new Timeframe with organized Interval(s) and respective values. It is the result of
// calling Flatten on Timeframe.All, and appending the resulting sequence to a new instance of Timeframe.
func (t *Timeframe[K, T]) Organize(cmp func(a, b T) bool, offset time.Duration) *Timeframe[K, T] {
	seq := Flatten(cmpFunc[K](cmp), mergeFunc[K, T], offset)(t.All())

	tf := NewTimeframe[K, T]()

	seq(func(interval Interval, t map[K]T) bool {
		_ = tf.Add(interval, t)

		return true
	})

	//slices.SortFunc(tf.Index.Keys, func(a, b Interval) int {
	//	return a.From.Compare(b.From)
	//})

	return tf
}

// Merge joins the intervals and respective values of the Timeframe tf into the Timeframe t,
// by extracting a SeqKV of the same items from tf using Timeframe.All, and adding them into
// Timeframe t using Timeframe.Append.
func (t *Timeframe[K, T]) Merge(tf *Timeframe[K, T]) (err error) {
	return t.Append(tf.All())
}
