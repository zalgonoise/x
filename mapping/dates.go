package mapping

import (
	"slices"
	"time"
)

type Timeframe[T any] struct {
	buffer []DataInterval[T]
}

// NewTimeframe creates a NewTimeframe of type T, with an 2D-map of Interval's to type T.
func NewTimeframe[T any]() *Timeframe[T] {
	return &Timeframe[T]{
		buffer: make([]DataInterval[T], 0, minAlloc),
	}
}

func (t *Timeframe[T]) init() *Timeframe[T] {
	if t == nil {
		return NewTimeframe[T]()
	}

	if t.buffer == nil {
		t.buffer = make([]DataInterval[T], 0, minAlloc)
	}

	return t
}

// Add joins the Interval i and its values to the Timeframe t, while ordering its
// previously inserted Interval(s) in the process.
func (t *Timeframe[T]) Add(i Interval, value T) bool {
	t.buffer = append(t.buffer, DataInterval[T]{Interval: i, Data: value})

	return true
}

// Append iterates through the input SeqKV and adds all intervals and respective values
// to the Timeframe t.
func (t *Timeframe[T]) Append(seq SeqKV[Interval, T]) bool {
	if !seq(t.Add) {
		return false
	}

	return true
}

// All returns an iterator over the values in the Timeframe,
// through the indexed Interval values ordered by their From time.Time value.
func (t *Timeframe[T]) All() SeqKV[Interval, T] {
	return func(yield func(Interval, T) bool) bool {
		for i := range t.buffer {
			if !yield(t.buffer[i].Interval, t.buffer[i].Data) {
				return false
			}
		}

		return true
	}
}

// Organize returns a new Timeframe with organized Interval(s) and respective values. It is the result of
// calling the input ReducerFunc (like Flatten or Replace) on Timeframe.All, and appending the resulting sequence
// to a new instance of Timeframe.
func (t *Timeframe[T]) Organize(reducer ReducerFunc[T]) *Timeframe[T] {
	seq := reducer(t.All())

	tf := NewTimeframe[T]()

	seq(tf.Add)

	slices.SortFunc(tf.buffer, func(a, b DataInterval[T]) int {
		return a.Interval.From.Compare(b.Interval.From)
	})

	return tf
}

type TimeframeStream[T comparable] struct {
	buffer    []DataInterval[T]
	mergeFunc func(a, b T) T
	offset    time.Duration
}

// NewTimeframeStream creates a TimeframeStream of type T, with an 2D-map of Interval's to type T,
// that automatically organizes itself each time that an item (or sequence of items) is(/are) added or appended
// to the TimeframeStream.
func NewTimeframeStream[T comparable](mergeFunc func(a, b T) T, offset time.Duration) *TimeframeStream[T] {
	if mergeFunc == nil {
		mergeFunc = func(a, b T) T { return b }
	}

	return &TimeframeStream[T]{
		buffer:    make([]DataInterval[T], 0, minAlloc),
		mergeFunc: mergeFunc,
		offset:    offset,
	}
}

func (t *TimeframeStream[T]) init() *TimeframeStream[T] {
	if t == nil {
		return NewTimeframeStream(func(a, b T) T { return b }, 0)
	}

	if t.buffer == nil {
		t.buffer = make([]DataInterval[T], 0, minAlloc)
	}

	if t.mergeFunc == nil {
		t.mergeFunc = func(a, b T) T { return b }
	}

	return t
}

// Add joins the Interval i and its values to the Timeframe t, while ordering its
// previously inserted Interval(s) in the process.
func (t *TimeframeStream[T]) Add(i Interval, value T) bool {
	t.buffer = resolveConflicts(i, value, t.buffer, split, t.mergeFunc, t.offset)

	slices.SortFunc(t.buffer, func(a, b DataInterval[T]) int {
		return a.Interval.From.Compare(b.Interval.From)
	})

	return true
}

// Append iterates through the input SeqKV and adds all intervals and respective values
// to the Timeframe t.
func (t *TimeframeStream[T]) Append(seq SeqKV[Interval, T]) bool {
	if !seq(t.Add) {
		return false
	}

	return true
}

// All returns an iterator over the values in the Timeframe,
// through the indexed Interval values ordered by their From time.Time value.
func (t *TimeframeStream[T]) All() SeqKV[Interval, T] {
	return func(yield func(Interval, T) bool) bool {
		for i := range t.buffer {
			if !yield(t.buffer[i].Interval, t.buffer[i].Data) {
				return false
			}
		}

		return true
	}
}

// Organize returns a new Timeframe with organized Interval(s) and respective values. It is the result of
// calling the input ReducerFunc (like Flatten or Replace) on Timeframe.All, and appending the resulting sequence
// to a new instance of Timeframe.
func (t *TimeframeStream[T]) Organize() *TimeframeStream[T] {
	return t
}
