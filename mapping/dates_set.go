package mapping

import "slices"

type TimeframeSet[T any] struct {
	buffer []DataInterval[T]
}

// NewTimeframeSet creates a NewTimeframeSet of type T, with an 2D-map of Interval's to type T.
func NewTimeframeSet[T any]() *TimeframeSet[T] {
	return &TimeframeSet[T]{
		buffer: make([]DataInterval[T], 0, minAlloc),
	}
}

func (t *TimeframeSet[T]) init() *TimeframeSet[T] {
	if t == nil {
		return NewTimeframeSet[T]()
	}

	if t.buffer == nil {
		t.buffer = make([]DataInterval[T], 0, minAlloc)
	}

	return t
}

// Add joins the Interval i and its values to the TimeframeSet t, while ordering its
// previously inserted Interval(s) in the process.
func (t *TimeframeSet[T]) Add(i Interval, value T) bool {
	t.buffer = append(t.buffer, DataInterval[T]{Interval: i, Data: value})

	return true
}

// Append iterates through the input SeqKV and adds all intervals and respective values
// to the TimeframeSet t.
func (t *TimeframeSet[T]) Append(seq SeqKV[Interval, T]) bool {
	if !seq(t.Add) {
		return false
	}

	return true
}

// All returns an iterator over the values in the TimeframeSet,
// through the indexed Interval values ordered by their From time.Time value.
func (t *TimeframeSet[T]) All() SeqKV[Interval, T] {
	return func(yield func(Interval, T) bool) bool {
		for i := range t.buffer {
			if !yield(t.buffer[i].Interval, t.buffer[i].Data) {
				return false
			}
		}

		return true
	}
}

// Organize returns a new TimeframeSet with organized Interval(s) and respective values. It is the result of
// calling the input ReducerFunc (like Flatten or Replace) on TimeframeSet.All, and appending the resulting sequence
// to a new instance of TimeframeSet.
func (t *TimeframeSet[T]) Organize(reducer ReducerFunc[T]) *TimeframeSet[T] {
	seq := reducer(t.All())

	tf := NewTimeframeSet[T]()

	seq(tf.Add)

	slices.SortFunc(tf.buffer, func(a, b DataInterval[T]) int {
		return a.Interval.From.Compare(b.Interval.From)
	})

	return tf
}
