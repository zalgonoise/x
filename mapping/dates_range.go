package mapping

type TimeframeRange[T any] struct {
	Keys   []Interval
	Values map[Interval]T
}

// NewTimeframeRange creates a TimeframeRange of type T, with an 2D-map of Interval's to type T.
func NewTimeframeRange[T any]() *TimeframeRange[T] {
	return &TimeframeRange[T]{
		Keys:   make([]Interval, 0, minAlloc),
		Values: make(map[Interval]T, minAlloc),
	}
}

func (t *TimeframeRange[T]) init() *TimeframeRange[T] {
	if t == nil {
		return NewTimeframeRange[T]()
	}

	if t.Keys == nil {
		t.Keys = make([]Interval, 0, minAlloc)
	}

	if t.Values == nil {
		t.Values = make(map[Interval]T, minAlloc)
	}

	return t
}

// Add joins the Interval i and its values to the TimeframeRange t, while ordering its
// previously inserted Interval(s) in the process.
func (t *TimeframeRange[T]) Add(i Interval, value T) bool {
	if _, ok := t.Values[i]; ok {
		t.Values[i] = value

		return false
	}

	t.Keys = append(t.Keys, i)
	t.Values[i] = value

	return true
}

// Append iterates through the input SeqKV and adds all intervals and respective values
// to the TimeframeRange t.
func (t *TimeframeRange[T]) Append(seq SeqKV[Interval, T]) error {
	if !seq(t.Add) {
		return ErrAppendFailed
	}

	return nil
}

// All returns an iterator over the values in the TimeframeRange,
// through the indexed Interval values ordered by their From time.Time value.
func (t *TimeframeRange[T]) All() SeqKV[Interval, T] {
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

// Organize returns a new TimeframeRange with organized Interval(s) and respective values. It is the result of
// calling the input ReducerFunc (like Flatten or Replace) on TimeframeRange.All, and appending the resulting sequence
// to a new instance of TimeframeRange.
func (t *TimeframeRange[T]) Organize(reducer ReducerFunc[T]) (*TimeframeRange[T], error) {
	seq, err := reducer(t.All())
	if err != nil {
		return nil, err
	}

	tf := NewTimeframeRange[T]()
	if err = tf.Append(seq); err != nil {
		return nil, err
	}

	return tf, nil
}
