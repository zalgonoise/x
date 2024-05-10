package mapping

// TimeframeReplacer stores values in intervals of time, as a 2D-map of Interval to a map of types K and T; by handling
// conflicts as replacement-updates.
type TimeframeReplacer[K comparable, T any] struct {
	Keys   []Interval
	Values map[Interval]map[K]T
}

// NewTimeframeReplacer creates a TimeframeReplacer of types K and T, with an 2D-map of Interval's to a map of types
// K and T.
func NewTimeframeReplacer[K comparable, T any]() *TimeframeReplacer[K, T] {
	return &TimeframeReplacer[K, T]{
		Keys:   make([]Interval, 0, minAlloc),
		Values: make(map[Interval]map[K]T, minAlloc),
	}
}

// Add joins the Interval i and its values to the TimeframeReplacer t, while ordering its
// previously inserted Interval(s) in the process.
func (t *TimeframeReplacer[K, T]) Add(i Interval, values map[K]T) bool {
	if _, ok := t.Values[i]; ok {
		t.Values[i] = values

		return false
	}

	t.Keys = append(t.Keys, i)
	t.Values[i] = values

	return true
}

// Append iterates through the input SeqKV and adds all intervals and respective values
// to the TimeframeReplacer t.
func (t *TimeframeReplacer[K, T]) Append(seq SeqKV[Interval, map[K]T]) error {
	if !seq(t.Add) {
		return ErrAppendFailed
	}

	return nil
}

// All returns an iterator over the values in the TimeframeReplacer,
// through the indexed Interval values ordered by their From time.Time value.
func (t *TimeframeReplacer[K, T]) All() SeqKV[Interval, map[K]T] {
	return func(yield func(Interval, map[K]T) bool) bool {
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

// Organize returns a new TimeframeReplacer with organized Interval(s) and respective values. It is the result of
// calling Replace on TimeframeReplacer.All, and appending the resulting sequence to a new instance of
// TimeframeReplacer.
func (t *TimeframeReplacer[K, T]) Organize() (*TimeframeReplacer[K, T], error) {
	seq, err := Replace(t.All())
	if err != nil {
		return nil, err
	}

	tf := NewTimeframeReplacer[K, T]()
	if err = tf.Append(seq); err != nil {
		return nil, err
	}

	return tf, nil
}

// Merge joins the intervals and respective values of the TimeframeReplacer tf into the TimeframeReplacer t,
// by extracting a SeqKV of the same items from tf using TimeframeReplacer.All, and adding them into
// Timeframe t using TimeframeReplacer.Append.
func (t *TimeframeReplacer[K, T]) Merge(tf *TimeframeReplacer[K, T]) (err error) {
	return t.Append(tf.All())
}
