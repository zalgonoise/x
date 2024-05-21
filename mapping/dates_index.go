package mapping

import (
	"errors"
)

const minAlloc = 64

var ErrAppendFailed = errors.New("failed to append to timeframe")

// TimeframeIndex stores values in intervals of time, as an Index of Interval and a map of type T.
type TimeframeIndex[T any] struct {
	Index *Index[Interval, T]
}

// NewTimeframeIndex creates a TimeframeIndex of type T, with an index on the Interval's From time.Time.
func NewTimeframeIndex[T any]() *TimeframeIndex[T] {
	return &TimeframeIndex[T]{
		Index: NewIndex[Interval, T](
			make(map[Interval]T),
			WithIndex[T](func(a, b Interval) int {
				return a.From.Compare(b.From)
			}),
			WithZero[Interval, T](*new(T)),
		),
	}
}

func (t *TimeframeIndex[T]) init() *TimeframeIndex[T] {
	if t == nil {
		return NewTimeframeIndex[T]()
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

// Add joins the Interval i and its values to the TimeframeIndex t, while ordering its
// previously inserted Interval(s) in the process.
func (t *TimeframeIndex[T]) Add(i Interval, value T) bool {
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
// to the TimeframeIndex t.
func (t *TimeframeIndex[T]) Append(seq SeqKV[Interval, T]) (err error) {
	if !seq(t.Add) {
		return ErrAppendFailed
	}

	return nil
}

// All returns an iterator over the values in the TimeframeIndex,
// through the indexed Interval values ordered by their From time.Time value.
func (t *TimeframeIndex[T]) All() SeqKV[Interval, T] {
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

// Organize returns a new TimeframeIndex with organized Interval(s) and respective values. It is the result of
// calling Flatten on Timeframe.All, and appending the resulting sequence to a new instance of TimeframeIndex.
func (t *TimeframeIndex[T]) Organize(reducer ReducerFunc[T]) *TimeframeIndex[T] {
	seq := reducer(t.All())

	tf := NewTimeframeIndex[T]()

	seq(func(interval Interval, t T) bool {
		_ = tf.Add(interval, t)

		return true
	})

	return tf
}

// Merge joins the intervals and respective values of the TimeframeIndex tf into the TimeframeIndex t,
// by extracting a SeqKV of the same items from tf using TimeframeIndex.All, and adding them into
// TimeframeIndex t using TimeframeIndex.Append.
func (t *TimeframeIndex[T]) Merge(tf *TimeframeIndex[T]) (err error) {
	return t.Append(tf.All())
}
