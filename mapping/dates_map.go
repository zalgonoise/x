package mapping

type TimeframeMap[T any] struct {
	Keys   []Interval
	Values map[Interval]T
}

// NewTimeframeMap creates a TimeframeMap of type T, that includes a map of Interval's to type T alongside its (sorted)
// keys.
func NewTimeframeMap[T any]() *TimeframeMap[T] {
	return &TimeframeMap[T]{
		Keys:   make([]Interval, 0, minAlloc),
		Values: make(map[Interval]T, minAlloc),
	}
}

func (t *TimeframeMap[T]) init() *TimeframeMap[T] {
	if t == nil {
		return NewTimeframeMap[T]()
	}

	if t.Keys == nil {
		t.Keys = make([]Interval, 0, minAlloc)
	}

	if t.Values == nil {
		t.Values = make(map[Interval]T, minAlloc)
	}

	return t
}

// Add joins the Interval i and its values to the TimeframeMap t, while ordering its
// previously inserted Interval(s) in the process.
func (t *TimeframeMap[T]) Add(i Interval, value T) bool {
	if _, ok := t.Values[i]; ok {
		t.Values[i] = value

		return false
	}

	t.Keys = append(t.Keys, i)
	t.Values[i] = value

	return true
}

// Append iterates through the input SeqKV and adds all intervals and respective values
// to the TimeframeMap t.
func (t *TimeframeMap[T]) Append(seq SeqKV[Interval, T]) error {
	if !seq(t.Add) {
		return ErrAppendFailed
	}

	return nil
}

// All returns an iterator over the values in the TimeframeMap,
// through the indexed Interval values ordered by their From time.Time value.
func (t *TimeframeMap[T]) All() SeqKV[Interval, T] {
	return func(yield func(Interval, T) bool) bool {
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
// calling the input ReducerFunc (like Flatten or Replace) on TimeframeMap.All, and appending the resulting sequence
// to a new instance of TimeframeMap.
func (t *TimeframeMap[T]) Organize(reducer ReducerFunc[T]) *TimeframeMap[T] {
	seq := reducer(t.All())

	tf := NewTimeframeMap[T]()

	seq(func(interval Interval, t T) bool {
		_ = tf.Add(interval, t)

		return true
	})

	return tf
}
