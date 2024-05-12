package mapping

import (
	"slices"

	"github.com/zalgonoise/cfg"
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

// ReducerFunc describes a strategy to apply to a sequence, that returns a (reduced) sequence and an error.
//
// This type can be used to apply different strategies in a map-reduce scenario, when working with Interval ranges.
//
// It is passed into functions like Organize, and implementations of it include Replace and Flatten.
type ReducerFunc[T any] func(SeqKV[Interval, T]) (SeqKV[Interval, T], error)

// DataInterval contains the data as Data for a specific Interval, as an isolated data structure used when caching
// Intervals and data as Seq of Interval and Data are organized.
type DataInterval[T any] struct {
	Data     T
	Interval Interval
}

// IntervalSet describes a single period of time after combining two (separate) intervals,
// where an item could contain the current value, the next value or both, for a specific Interval.
type IntervalSet struct {
	cur  bool
	next bool
	i    Interval
}

// TimeframeType is a generic interface that describes a constructor to a timeframe implementation.
//
// The generic interface includes the init method, allowing for a generic function to safely create a
// pointer to type M as required by the type itself.
//
// Other than initializing type M, the type simply needs to ingest a sequence of Interval and type T,
// and that is where Append comes in.
type TimeframeType[T any, M any] interface {
	*M
	init() *M
	Add(i Interval, value T) bool
}

func AsSeq[T any](data []DataInterval[T]) SeqKV[Interval, T] {
	return func(yield func(Interval, T) bool) bool {
		for i := range data {
			if !yield(data[i].Interval, data[i].Data) {
				return false
			}
		}

		return true
	}
}

// Organize consumes a sequence Seq of Interval and data, returning an appropriate timeframe type K with organized
// / flattened values.
func Organize[M TimeframeType[T, K], T any, K any](seq SeqKV[Interval, T], reducer ReducerFunc[T]) (*K, error) {
	flattened, err := reducer(seq)
	if err != nil {
		return nil, err
	}

	var tf M = new(K)
	tf = tf.init()

	flattened(func(interval Interval, t T) bool {
		_ = tf.Add(interval, t)

		return true
	})

	return tf, nil
}

// Replace consumes the sequence Seq of Interval and data to align all From and To time.Time values for each Interval
// and coalescing the data in any intersections when required. It returns a similar but new instance of the same type of
// Seq, but with organized values.
func Replace[T any]() ReducerFunc[T] {
	return func(seq SeqKV[Interval, T]) (reduced SeqKV[Interval, T], err error) {
		cache := make([]DataInterval[T], 0, minAlloc)

		if !seq(func(interval Interval, value T) bool {
			conflicts, indices := findConflicts(cache, interval)

			if len(conflicts) == 0 {
				cache = append(cache, DataInterval[T]{
					Data:     value,
					Interval: interval,
				})

				return true
			}

			// remove conflicts from cache
			for i := len(indices) - 1; i >= 0; i-- {
				cache = slices.Delete(cache, indices[i], indices[i]+1)
			}

			for i := range conflicts {
				sets, overlaps, err := replace(conflicts[i].Interval, interval)
				if err != nil {
					return false
				}

				if !overlaps {
					continue
				}

				for idx := range sets {
					switch {
					case sets[idx].cur && !sets[idx].next:
						cache = append(cache, DataInterval[T]{
							Data:     conflicts[i].Data,
							Interval: sets[idx].i,
						})
					case !sets[idx].cur && sets[idx].next:
						cache = append(cache, DataInterval[T]{
							Data:     value,
							Interval: sets[idx].i,
						})
					default:
						// unsupported state in Replace
						continue
					}
				}
			}

			return true
		}) {
			return nil, err
		}

		return func(yield func(Interval, T) bool) bool {
			for i := range cache {
				if !yield(cache[i].Interval, cache[i].Data) {
					return false
				}
			}

			return true
		}, nil
	}
}

// Flatten consumes the sequence Seq of Interval and data to align all From and To time.Time values for each Interval
// and coalescing the data in any intersections when required. It returns a similar but new instance of the same type of
// Seq, but with organized values.
func Flatten[T any](cmpFunc func(a, b T) bool, mergeFunc func(a, b T) T) ReducerFunc[T] {
	return func(seq SeqKV[Interval, T]) (reduced SeqKV[Interval, T], err error) {
		cache := make([]DataInterval[T], 0, minAlloc)

		if !seq(func(interval Interval, value T) bool {
			cache = mergeCache(cache, cmpFunc)

			conflicts, indices := findConflicts(cache, interval)

			if len(conflicts) == 0 {
				cache = append(cache, DataInterval[T]{
					Data:     value,
					Interval: interval,
				})

				return true
			}

			// remove conflicts from cache
			for i := len(indices) - 1; i >= 0; i-- {
				cache = slices.Delete(cache, indices[i], indices[i]+1)
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
						cache = append(cache, DataInterval[T]{
							Data:     conflicts[i].Data,
							Interval: sets[idx].i,
						})
					case !sets[idx].cur && sets[idx].next:
						cache = append(cache, DataInterval[T]{
							Data:     value,
							Interval: sets[idx].i,
						})
					default:
						cache = append(cache, DataInterval[T]{
							Data:     mergeFunc(conflicts[i].Data, value),
							Interval: sets[idx].i,
						})
					}
				}
			}

			return true
		}) {
			return nil, err
		}

		return func(yield func(Interval, T) bool) bool {
			for i := range cache {
				if !yield(cache[i].Interval, cache[i].Data) {
					return false
				}
			}

			return true
		}, nil
	}
}

func FormatTime[T any](
	seq SeqKV[Interval, T],
	opts ...cfg.Option[Format],
) SeqKV[Interval, T] {
	format := cfg.New(opts...)

	if format.fnFrom == nil && format.fnTo == nil {
		return seq
	}

	return func(yield func(Interval, T) bool) bool {
		return seq(func(interval Interval, m T) bool {
			var ok bool

			if format.fnFrom != nil {
				interval.From, ok = format.fnFrom(interval.From)
				if !ok {
					return false
				}
			}

			if format.fnTo != nil {
				interval.To, ok = format.fnTo(interval.To)
				if !ok {
					return false
				}
			}

			return yield(interval, m)
		})
	}
}

func mergeCache[T any](cache []DataInterval[T], cmp func(a, b T) bool) []DataInterval[T] {
	if len(cache) <= 1 {
		return cache
	}

	for i := len(cache) - 1; i > 0; i-- {
		if cache[i-1].Interval.To.Equal(cache[i].Interval.From) && cmp(cache[i-1].Data, cache[i].Data) {
			cache[i-1].Interval.To = cache[i].Interval.To
			cache = slices.Delete(cache, i, i+1)
		}
	}

	return cache
}

func findConflicts[T any](cache []DataInterval[T], cur Interval) (conflicts []DataInterval[T], indices []int) {
	conflicts = make([]DataInterval[T], 0, len(cache))
	indices = make([]int, 0, len(cache))

	for i := range cache {
		switch {
		case cache[i].Interval.To.Compare(cur.From) <= 0, cache[i].Interval.From.Compare(cur.To) >= 0:
			continue
		default:
			conflicts = append(conflicts, cache[i])
			indices = append(indices, i)
		}
	}

	return conflicts, indices
}

func replace(cur, next Interval) ([]IntervalSet, bool, error) {
	switch {
	// next is after
	//  c -  |###|
	//  n -        |###|
	case cur.To.Before(next.From):
		return []IntervalSet{{cur: true, i: cur}, {next: true, i: next}}, false, nil

	// cur is after
	//  c -        |###|
	//  n -  |###|
	case cur.From.After(next.To):
		return []IntervalSet{{next: true, i: next}, {cur: true, i: cur}}, false, nil

	// overlapping start
	//  c -  |#######|
	//  n -  |#####?????
	case cur.From.Equal(next.From):
		switch cur.To.Compare(next.To) {
		case -1:
			// before
			//  c -  |#######|
			//  n -  |###########|
			return []IntervalSet{
				{next: true, i: Interval{From: next.From, To: next.To}},
			}, true, nil
		case 1:
			// after
			//  c -  |#######|
			//  n -  |#####|
			return []IntervalSet{
				{next: true, i: Interval{From: next.From, To: next.To}},
				{cur: true, i: Interval{From: next.To, To: cur.To}},
			}, true, nil
		case 0:
			// equal
			//  c -  |#######|
			//  n -  |#######|
			return []IntervalSet{
				{next: true, i: Interval{From: next.From, To: next.To}},
			}, true, nil
		}

	// overlap: portion of end
	//  c -  |#######|
	//  n -     |##?????
	case next.From.After(cur.From):
		switch next.To.Compare(cur.To) {
		case 1:
			// after; next goes beyond cur
			//  c -  |#######|
			//  n -     |#######|
			return []IntervalSet{
				{cur: true, i: Interval{From: cur.From, To: next.From}},
				{next: true, i: Interval{From: next.From, To: next.To}},
			}, true, nil
		case -1:
			// before; next is within cur
			//  c -  |#######|
			//  n -     |###|
			return []IntervalSet{
				{cur: true, i: Interval{From: cur.From, To: next.From}},
				{next: true, i: Interval{From: next.From, To: next.To}},
				{cur: true, i: Interval{From: next.To, To: cur.To}},
			}, true, nil
		case 0:
			// matching ends
			//  c -  |#######|
			//  n -     |####|
			return []IntervalSet{
				{cur: true, i: Interval{From: cur.From, To: next.From}},
				{next: true, i: Interval{From: next.From, To: next.To}},
			}, true, nil
		}

	// overlap: portion of start
	//  c -    |#######|
	//  n -  |##?????
	case next.From.Before(cur.From):
		switch next.To.Compare(cur.To) {
		case 1:
			// overlap: entirety of current
			//  c -    |#######|
			//  n -  |############|
			return []IntervalSet{
				{next: true, i: Interval{From: next.From, To: next.To}},
			}, true, nil
		case -1:
			// overlap: portion of start
			//  c -    |#######|
			//  n -  |######|
			return []IntervalSet{
				{next: true, i: Interval{From: next.From, To: next.To}},
				{cur: true, i: Interval{From: next.To, To: cur.To}},
			}, true, nil
		case 0:
			// overlap: entirety of current
			//  c -    |#######|
			//  n -  |#########|
			return []IntervalSet{
				{next: true, i: Interval{From: next.From, To: next.To}},
			}, true, nil
		}
	}

	return nil, false, ErrTimeSplitFailed
}

func split(cur, next Interval) ([]IntervalSet, bool, error) {
	switch {
	// next is after
	//  c -  |###|
	//  n -        |###|
	case cur.To.Before(next.From):
		return []IntervalSet{{cur: true, i: cur}, {next: true, i: next}}, false, nil

	// cur is after
	//  c -        |###|
	//  n -  |###|
	case cur.From.After(next.To):
		return []IntervalSet{{next: true, i: next}, {cur: true, i: cur}}, false, nil

	// overlapping start
	//  c -  |#######|
	//  n -  |#####?????
	case cur.From.Equal(next.From):
		switch cur.To.Compare(next.To) {
		case -1:
			// before
			//  c -  |#######|
			//  n -  |###########|
			return []IntervalSet{
				{cur: true, next: true, i: Interval{From: cur.From, To: cur.To}},
				{next: true, i: Interval{From: cur.To, To: next.To}},
			}, true, nil
		case 1:
			// after
			//  c -  |#######|
			//  n -  |#####|
			return []IntervalSet{
				{cur: true, next: true, i: Interval{From: cur.From, To: next.To}},
				{cur: true, i: Interval{From: next.To, To: cur.To}},
			}, true, nil
		case 0:
			// equal
			//  c -  |#######|
			//  n -  |#######|
			return []IntervalSet{
				{cur: true, next: true, i: cur},
			}, true, nil
		}

	// overlap: portion of end
	//  c -  |#######|
	//  n -     |##?????
	case next.From.After(cur.From):
		switch next.To.Compare(cur.To) {
		case 1:
			// after; next goes beyond cur
			//  c -  |#######|
			//  n -     |#######|
			return []IntervalSet{
				{cur: true, i: Interval{From: cur.From, To: next.From}},
				{cur: true, next: true, i: Interval{From: next.From, To: cur.To}},
				{next: true, i: Interval{From: cur.To, To: next.To}},
			}, true, nil
		case -1:
			// before; next is within cur
			//  c -  |#######|
			//  n -     |###|
			return []IntervalSet{
				{cur: true, i: Interval{From: cur.From, To: next.From}},
				{cur: true, next: true, i: Interval{From: next.From, To: next.To}},
				{cur: true, i: Interval{From: next.To, To: cur.To}},
			}, true, nil
		case 0:
			// matching ends
			//  c -  |#######|
			//  n -     |####|
			return []IntervalSet{
				{cur: true, i: Interval{From: cur.From, To: next.From}},
				{cur: true, next: true, i: Interval{From: next.From, To: next.To}},
			}, true, nil
		}

	// overlap: portion of start
	//  c -    |#######|
	//  n -  |##?????
	case next.From.Before(cur.From):
		switch next.To.Compare(cur.To) {
		case 1:
			// overlap: entirety of current
			//  c -    |#######|
			//  n -  |############|
			return []IntervalSet{
				{next: true, i: Interval{From: next.From, To: cur.From}},
				{cur: true, next: true, i: Interval{From: cur.From, To: cur.To}},
				{next: true, i: Interval{From: cur.To, To: next.To}},
			}, true, nil
		case -1:
			// overlap: portion of start
			//  c -    |#######|
			//  n -  |######|
			return []IntervalSet{
				{next: true, i: Interval{From: next.From, To: cur.From}},
				{cur: true, next: true, i: Interval{From: cur.From, To: next.To}},
				{cur: true, i: Interval{From: next.To, To: cur.To}},
			}, true, nil
		case 0:
			// overlap: entirety of current
			//  c -    |#######|
			//  n -  |#########|
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
