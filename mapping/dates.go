package mapping

import (
	"errors"
	"time"
)

const minAlloc = 64

var ErrAppendFailed = errors.New("failed to append to timeframe")

// Timeframe stores values in intervals of time, as an Index of Interval and a map of types K and T.
type Timeframe[T any] struct {
	Index *Index[Interval, T]
}

// Interval is a period of time with a From and To time.Time values.
type Interval struct {
	From time.Time
	To   time.Time
}

// NewTimeframe creates a Timeframe of types K and T, with an index on the Interval's From time.Time.
func NewTimeframe[T any]() *Timeframe[T] {
	return &Timeframe[T]{
		Index: NewIndex[Interval, T](
			make(map[Interval]T),
			WithIndex[T](func(a, b Interval) int {
				return a.From.Compare(b.From)
			}),
			WithZero[Interval, T](*new(T)),
		),
	}
}

func (t *Timeframe[T]) init() *Timeframe[T] {
	if t == nil {
		return NewTimeframe[T]()
	}

	if t.Index == nil {
		t.Index = NewIndex[Interval, T](
			make(map[Interval]T),
			WithIndex[T](func(a, b Interval) int {
				return a.From.Compare(b.From)
			}),
			WithZero[Interval, T](*new(T)),
		)
	}

	return t
}

// Add joins the Interval i and its values to the Timeframe t, while ordering its
// previously inserted Interval(s) in the process.
func (t *Timeframe[T]) Add(i Interval, value T) bool {
	if i.To.Before(i.From) {
		return false
	}

	if len(t.Index.Keys) == 0 {
		t.Index.Set(i, func(old T) (T, bool) {
			return value, true
		})

		return true
	}

	if _, ok := t.Index.Get(i); ok {
		t.Index.Set(i, func(old T) (T, bool) {
			return value, true
		})

		return true
	}

	t.Index.Set(i, func(old T) (T, bool) {
		return value, true
	})

	return true
}

// Append iterates through the input SeqKV and adds all intervals and respective values
// to the Timeframe t.
func (t *Timeframe[T]) Append(seq SeqKV[Interval, T]) (err error) {
	if !seq(t.Add) {
		return ErrAppendFailed
	}

	return nil
}

// All returns an iterator over the values in the Timeframe,
// through the indexed Interval values ordered by their From time.Time value.
func (t *Timeframe[T]) All() SeqKV[Interval, T] {
	return func(yield func(Interval, T) bool) bool {
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
func (t *Timeframe[T]) Organize(reducer ReducerFunc[T]) *Timeframe[T] {
	seq := reducer(t.All())

	tf := NewTimeframe[T]()

	seq(func(interval Interval, t T) bool {
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
func (t *Timeframe[T]) Merge(tf *Timeframe[T]) (err error) {
	return t.Append(tf.All())
}
