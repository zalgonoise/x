package mapping

import (
	"maps"
	"time"
)

// SeqKV describes a sequence of iterable items, which takes a yield func which will be used
// to perform a certain operation on each yielded item throughout the iteration
//
// ref: https://github.com/golang/go/issues/61899
type SeqKV[T, K any] func(yield func(T, K) bool) bool

// Seq describes a sequence of iterable items, which takes a yield func which will be used
// to perform a certain operation on each yielded item throughout the iteration
//
// ref: https://github.com/golang/go/issues/61899
type Seq[T any] func(yield func(T) bool) bool

// DataInterval contains the data as Data for a specific Interval, as an isolated data structure used when caching
// Intervals and data as Seq of Interval and Data are organized.
type DataInterval[K comparable, T any] struct {
	Data     map[K]T
	Interval Interval
}

// IntervalSet describes a single period of time after combining two (separate) intervals,
// where an item could contain the current value, the next value or both, for a specific Interval.
type IntervalSet struct {
	cur  bool
	next bool
	i    Interval
}

// Organize consumes a sequence Seq of Interval and data, returning a Timeframe with organized / flattened values.
func Organize[K comparable, T any](seq SeqKV[Interval, map[K]T]) (tf *Timeframe[K, T], err error) {
	flattened, err := Flatten(seq)
	if err != nil {
		return nil, err
	}

	tf = NewTimeframe[K, T]()
	if err = tf.Append(flattened); err != nil {
		return nil, err
	}

	return tf, nil
}

// OrganizeMap consumes a sequence Seq of Interval and data, returning a TimeframeMap with organized / flattened values.
func OrganizeMap[K comparable, T any](seq SeqKV[Interval, map[K]T]) (tf *TimeframeMap[K, T], err error) {
	flattened, err := Flatten(seq)
	if err != nil {
		return nil, err
	}

	tf = NewTimeframeMap[K, T]()
	if err = tf.Append(flattened); err != nil {
		return nil, err
	}

	return tf, nil
}

// Flatten consumes the sequence Seq of Interval and data to align all From and To time.Time values for each Interval
// and coalescing the data in any intersections when required. It returns a similar but new instance of the same type of
// Seq, but with organized values.
func Flatten[K comparable, T any](seq SeqKV[Interval, map[K]T]) (sorted SeqKV[Interval, map[K]T], err error) {
	cache := make([]DataInterval[K, T], 0, minAlloc)

	if !seq(func(interval Interval, m map[K]T) bool {
		if len(cache) == 0 {
			cache = append(cache, DataInterval[K, T]{
				Data:     m,
				Interval: interval,
			})

			return true
		}

		conflicts := findConflicts(cache, interval)

		if len(conflicts) == 0 {
			cache = append(cache, DataInterval[K, T]{
				Data:     m,
				Interval: interval,
			})

			return true
		}

		for i := range conflicts {
			sets, overlaps, err := split(conflicts[i].Interval, interval)
			if err != nil {
				return false
			}

			if !overlaps {
				continue
			}

			for idx := range sets {
				switch {
				case sets[idx].cur && !sets[idx].next:
					cache = append(cache, DataInterval[K, T]{
						Data:     conflicts[i].Data,
						Interval: sets[idx].i,
					})
				case !sets[idx].cur && sets[idx].next:
					cache = append(cache, DataInterval[K, T]{
						Data:     m,
						Interval: sets[idx].i,
					})
				default:
					dataCopy := maps.Clone(m)
					coalesce(dataCopy, conflicts[i].Data)

					cache = append(cache, DataInterval[K, T]{
						Data:     dataCopy,
						Interval: sets[idx].i,
					})
				}
			}
		}

		return true
	}) {
		return nil, err
	}

	return func(yield func(Interval, map[K]T) bool) bool {
		for i := range cache {
			if !yield(cache[i].Interval, cache[i].Data) {
				return false
			}
		}

		return true
	}, nil
}

func FormatTime[K comparable, T any](
	seq SeqKV[Interval, map[K]T],
	fnFrom func(time.Time) (time.Time, bool),
	fnTo func(time.Time) (time.Time, bool),
) SeqKV[Interval, map[K]T] {
	if fnFrom == nil && fnTo == nil {
		return seq
	}

	return func(yield func(Interval, map[K]T) bool) bool {
		return seq(func(interval Interval, m map[K]T) bool {
			var ok bool

			if fnFrom != nil {
				interval.From, ok = fnFrom(interval.From)
				if !ok {
					return false
				}
			}

			if fnTo != nil {
				interval.To, ok = fnTo(interval.To)
				if !ok {
					return false
				}
			}

			return yield(interval, m)
		})
	}
}

func findConflicts[K comparable, T any](cache []DataInterval[K, T], cur Interval) []DataInterval[K, T] {
	conflicts := make([]DataInterval[K, T], 0, len(cache))

	for i := range cache {
		switch {
		case cache[i].Interval.To.Compare(cur.From) <= 0, cache[i].Interval.From.Compare(cur.To) >= 0:
			continue
		default:
			conflicts = append(conflicts, cache[i])
		}
	}

	return conflicts
}

func split(cur, next Interval) ([]IntervalSet, bool, error) {
	switch {
	// next is after
	case cur.To.Before(next.From):
		return []IntervalSet{{cur: true, i: cur}, {next: true, i: next}}, false, nil

	// cur is after
	case cur.From.After(next.To):
		return []IntervalSet{{next: true, i: next}, {cur: true, i: cur}}, false, nil

	// overlapping start
	case cur.From.Equal(next.From):
		switch cur.To.Compare(next.To) {
		case -1: // before
			return []IntervalSet{
				{cur: true, next: true, i: Interval{From: cur.From, To: cur.To}},
				{next: true, i: Interval{From: cur.To, To: next.To}},
			}, true, nil
		case 1: // after
			return []IntervalSet{
				{cur: true, next: true, i: Interval{From: cur.From, To: next.To}},
				{cur: true, i: Interval{From: next.To, To: cur.To}},
			}, true, nil
		case 0: // equal
			return []IntervalSet{
				{cur: true, next: true, i: cur},
			}, true, nil
		}

	// overlap: portion of end
	case next.From.After(cur.From):
		switch next.To.Compare(cur.To) {
		case 1: // after; next goes beyond cur
			return []IntervalSet{
				{cur: true, i: Interval{From: cur.From, To: next.From}},
				{cur: true, next: true, i: Interval{From: next.From, To: cur.To}},
				{next: true, i: Interval{From: cur.To, To: next.To}},
			}, true, nil
		case -1: // before; next is within cur
			return []IntervalSet{
				{cur: true, i: Interval{From: cur.From, To: next.From}},
				{cur: true, next: true, i: Interval{From: next.From, To: next.To}},
				{cur: true, i: Interval{From: next.To, To: cur.To}},
			}, true, nil
		case 0:
			return []IntervalSet{
				{cur: true, i: Interval{From: cur.From, To: next.From}},
				{cur: true, next: true, i: Interval{From: next.From, To: next.To}},
			}, true, nil
		}
	case next.From.Before(cur.From):
		switch next.To.Compare(cur.To) {
		case 1:
			return []IntervalSet{
				{next: true, i: Interval{From: next.From, To: cur.From}},
				{cur: true, next: true, i: Interval{From: cur.From, To: cur.To}},
				{next: true, i: Interval{From: cur.To, To: next.To}},
			}, true, nil
		case -1:
			return []IntervalSet{
				{next: true, i: Interval{From: next.From, To: cur.From}},
				{cur: true, next: true, i: Interval{From: cur.From, To: next.To}},
				{cur: true, i: Interval{From: next.To, To: cur.To}},
			}, true, nil
		case 0:
			return []IntervalSet{
				{next: true, i: Interval{From: next.From, To: cur.From}},
				{cur: true, next: true, i: Interval{From: cur.From, To: cur.To}},
			}, true, nil
		}
	}

	return nil, false, ErrTimeSplitFailed
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
