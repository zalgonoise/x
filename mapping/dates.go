package mapping

import (
	"errors"
	"maps"
	"time"
)

const defaultMaxGap = time.Minute

var (
	errZeroOrNegativeDur     = errors.New("interval cannot end at the same time or before start")
	errOverlappingTimeframes = errors.New("overlapping timeframes")
	errGapBetweenTimeframes  = errors.New("gap between timeframes")
	errAppendFailed          = errors.New("failed to append to timeframe")
	errTimeSplitFailed       = errors.New("failed to split time intervals")
)

type Interval struct {
	From time.Time
	To   time.Time
}

type Timeframe[K comparable, T any] struct {
	noOverlap bool
	//maxGap    time.Duration

	Index *Index[Interval, map[K]T]
}

type IntervalSet struct {
	cur  bool
	next bool
	i    Interval
}

// Add joins the Interval i and its values to the Timeframe t, while ordering its
// previously inserted Interval(s) in the process.
func (t *Timeframe[K, T]) Add(i Interval, values map[K]T) error {
	if i.To.Before(i.From) {
		return errZeroOrNegativeDur
	}

	if len(t.Index.Keys) == 0 {
		t.Index.Set(i, func(old map[K]T) (map[K]T, bool) {
			return values, true
		})

		return nil
	}

	// TODO: what if we're adding an interval way back in the past?
	//
	// needs to scan the entire indexed keys to match any overlaps within the
	// Interval i's From and To values
	//
	// TestTimeframe --> interleaved/recursive: reproduces this event
	//
	// this would become more and more expensive the more passes we do through the data
	// so organizing the intervals separately seems best. implemented within this function given the current logic,
	// and certain keys and values get lost in the end. That is not ideal.
	//
	// for that we can split the intervals like it's done here (as it works) but in a new Index,
	// and collecting the data into a set, concatenating it in the end.
	last := t.Index.Keys[len(t.Index.Keys)-1]
	lastVal := t.Index.values[last]

	// iterative process to join intervals if needed; should cycle through
	// all occurrences where current interval overlaps
	sets, err := t.split(last, i)
	if err != nil {
		return err
	}

	switch {
	case len(sets) == 1 && sets[0].cur && sets[0].next:
		t.Index.values[last] = coalesce(lastVal, values)
	// keep last as-is
	case len(sets) == 2 && sets[0].i == last:
		t.Index.Set(i, func(old map[K]T) (map[K]T, bool) {
			return values, true
		})
	default:
		t.Index.Keys = t.Index.Keys[:len(t.Index.Keys)-1]
		delete(t.Index.values, last)

		for idx := range sets {
			switch {
			case sets[idx].cur && sets[idx].next:
				valuesCopy := make(map[K]T, len(values))
				maps.Copy(valuesCopy, values)

				t.Index.Set(sets[idx].i, func(old map[K]T) (map[K]T, bool) {
					return coalesce(valuesCopy, lastVal), true
				})
			case sets[idx].cur && !sets[idx].next:
				t.Index.Set(sets[idx].i, func(old map[K]T) (map[K]T, bool) {
					return lastVal, true
				})
			case !sets[idx].cur && sets[idx].next:
				t.Index.Set(sets[idx].i, func(old map[K]T) (map[K]T, bool) {
					return values, true
				})
			}
		}
	}

	return nil
}

// TODO: needs rework with a different strategy, in order to flatten any type of Interval combination:
//  1. iterate through all Interval values, and create an ordered slice of their From time.Time values, along with the
//     index to the entry they are pointing to (that is starting)
//  2. iterate through all Interval values, and create an ordered slice of their To time.Time values, along with the
//     index to the entry they are pointing to (that is ending)
//     -- this can be combined with step 1, where the indices are shared between starting and ending slices
//  3. iterate through all From time.Time values, and find the following To time.Time value. This is the first reference
//     point to that particular interval
//     Then, iterate through the next From time.Time values to find another From value before the next To. This allows
//     splitting Intervals according to their content; where the involved indices to the data can be referenced and
//     aggregated if applicable.
//  4. with the final Intervals and data indices, rebuild or initialize a new *Timeframe type with the flattened
//     Intervals, and the aggregated data that they should have respectively.
//     ---
//     Ideally, this should be done separately from the Add method, to avoid multiple iterations on the underlying
//     Timeframe Interval values. The All and Append methods can stay, however, with a similar sequence / iter approach
//
// split takes two Interval values, returning a slice of IntervalSet and an error.
//
// This method provides context to the caller on how many data points and their respective intervals of time and data,
// when combining Timeframe Interval values.
func (t *Timeframe[K, T]) split(cur, next Interval) ([]IntervalSet, error) {
	switch {
	// next is after
	case cur.To.Before(next.From):
		//if next.From.Sub(cur.To) > t.maxGap {
		//	return nil, errGapBetweenTimeframes
		//}

		return []IntervalSet{{cur: true, i: cur}, {next: true, i: next}}, nil

	// cur is after
	case cur.From.After(next.To):
		//if cur.From.Sub(next.To) > t.maxGap {
		//	return nil, errGapBetweenTimeframes
		//}

		return []IntervalSet{{next: true, i: next}, {cur: true, i: cur}}, nil

	// overlapping start
	case cur.From.Equal(next.From):
		if t.noOverlap {
			return nil, errOverlappingTimeframes
		}

		switch cur.To.Compare(next.To) {
		case -1: // before
			return []IntervalSet{
				{cur: true, next: true, i: Interval{From: cur.From, To: cur.To}},
				{next: true, i: Interval{From: cur.To, To: next.To}},
			}, nil
		case 1: // after
			return []IntervalSet{
				{cur: true, next: true, i: Interval{From: cur.From, To: next.To}},
				{cur: true, i: Interval{From: next.To, To: cur.To}},
			}, nil
		case 0: // equal
			return []IntervalSet{
				{cur: true, next: true, i: cur},
			}, nil
		}

	// overlap: portion of end
	case next.From.After(cur.From):
		if t.noOverlap {
			return nil, errOverlappingTimeframes
		}

		switch next.To.Compare(cur.To) {
		case 1: // after; next goes beyond cur
			return []IntervalSet{
				{cur: true, i: Interval{From: cur.From, To: next.From}},
				{cur: true, next: true, i: Interval{From: next.From, To: cur.To}},
				{next: true, i: Interval{From: cur.To, To: next.To}},
			}, nil
		case -1: // before; next is within cur
			return []IntervalSet{
				{cur: true, i: Interval{From: cur.From, To: next.From}},
				{cur: true, next: true, i: Interval{From: next.From, To: next.To}},
				{next: true, i: Interval{From: next.To, To: cur.To}},
			}, nil
		case 0:
			return []IntervalSet{
				{cur: true, i: Interval{From: cur.From, To: next.From}},
				{cur: true, next: true, i: Interval{From: next.From, To: next.To}},
			}, nil
		}

	// overlap: portion of start
	case next.From.Before(cur.From):
		switch next.To.Compare(cur.To) {
		case 1:
			return []IntervalSet{
				{next: true, i: Interval{From: next.From, To: cur.From}},
				{cur: true, next: true, i: Interval{From: cur.From, To: cur.To}},
				{next: true, i: Interval{From: cur.To, To: next.To}},
			}, nil
		case -1:
			return []IntervalSet{
				{next: true, i: Interval{From: next.From, To: cur.From}},
				{cur: true, next: true, i: Interval{From: cur.From, To: next.To}},
				{cur: true, i: Interval{From: next.To, To: cur.To}},
			}, nil
		case 0:
			return []IntervalSet{
				{next: true, i: Interval{From: next.From, To: cur.From}},
				{cur: true, next: true, i: Interval{From: cur.From, To: cur.To}},
			}, nil
		}
	}

	return nil, errTimeSplitFailed
}

// Seq describes a sequence of iterable items, which takes a yield func which will be used
// to perform a certain operation on each yielded item throughout the iteration
//
// ref: https://github.com/golang/go/issues/61899
type Seq[T, K any] func(yield func(T, K) bool) bool

// All returns an iterator over the values in the Timeframe,
// through the indexed Interval values ordered by their From time.Time value.
func (t *Timeframe[K, T]) All() Seq[Interval, map[K]T] {
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

// Append iterates through the input Seq and adds all intervals and respective values
// to the Timeframe t.
func (t *Timeframe[K, T]) Append(seq Seq[Interval, map[K]T]) (err error) {
	if !seq(func(interval Interval, m map[K]T) bool {
		if err = t.Add(interval, m); err != nil {
			return false
		}

		return true
	}) {
		if err != nil {
			return err
		}

		return errAppendFailed
	}

	return nil
}

// Merge joins the intervals and respective values of the Timeframe tf into the Timeframe t,
// by extracting a Seq of the same items from tf using Timeframe.All, and adding them into
// Timeframe t using Timeframe.Append.
func (t *Timeframe[K, T]) Merge(tf *Timeframe[K, T]) (err error) {
	return t.Append(tf.All())
}

func NewTimeframe[K comparable, T any]() *Timeframe[K, T] {
	return &Timeframe[K, T]{
		//maxGap: defaultMaxGap,
		Index: NewIndex[Interval, map[K]T](
			make(map[Interval]map[K]T),
			WithIndex[map[K]T](func(a, b Interval) int {
				return a.From.Compare(b.From)
			}),
			WithZero[Interval, map[K]T](nil),
		),
	}
}

func coalesce[K comparable, T any](start, next map[K]T) map[K]T {
	switch {
	case start != nil && next == nil:
		return start
	case start == nil && next != nil:
		return next
	case start == nil && next == nil:
		return nil
	}

	for key, value := range next {
		start[key] = value
	}

	return start
}

func coalesceUnset[K comparable, T any](start, next map[K]T) (res map[K]T, skipped []K) {
	switch {
	case start != nil && next == nil:
		return start, nil
	case start == nil && next != nil:
		return next, nil
	case start == nil && next == nil:
		return nil, nil
	}

	skipped = make([]K, 0, len(next))

	for key, value := range next {
		if _, ok := start[key]; ok {
			skipped = append(skipped, key)

			continue
		}

		start[key] = value
	}

	if len(skipped) == 0 {
		return start, nil
	}

	return start, skipped
}
