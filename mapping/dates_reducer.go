package mapping

import (
	"slices"
	"time"
)

// ReducerFunc describes a strategy to apply to a sequence, that returns a (reduced) sequence and an error.
//
// This type can be used to apply different strategies in a map-reduce scenario, when working with Interval ranges.
//
// It is passed into functions like Organize, and implementations of it include Replace and Flatten.
type ReducerFunc[T any] func(SeqKV[Interval, T]) SeqKV[Interval, T]

// Replace consumes the sequence Seq of Interval and data to align all From and To time.Time values for each Interval
// and coalescing the data in any intersections when required. It returns a similar but new instance of the same type of
// Seq, but with organized values.
func Replace[T any](offset time.Duration) ReducerFunc[T] {
	return func(seq SeqKV[Interval, T]) (reduced SeqKV[Interval, T]) {
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
				sets, overlaps := replace(conflicts[i].Interval, interval, offset)

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
			return nil
		}

		return func(yield func(Interval, T) bool) bool {
			for i := range cache {
				if !yield(cache[i].Interval, cache[i].Data) {
					return false
				}
			}

			return true
		}
	}
}

// Flatten consumes the sequence Seq of Interval and data to align all From and To time.Time values for each Interval
// and coalescing the data in any intersections when required. It returns a similar but new instance of the same type of
// Seq, but with organized values.
func Flatten[T any](cmpFunc func(a, b T) bool, mergeFunc func(a, b T) T, offset time.Duration) ReducerFunc[T] {
	return func(seq SeqKV[Interval, T]) (reduced SeqKV[Interval, T]) {
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
				sets, overlaps := split(conflicts[i].Interval, interval, offset)

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
			return nil
		}

		return func(yield func(Interval, T) bool) bool {
			for i := range cache {
				if !yield(cache[i].Interval, cache[i].Data) {
					return false
				}
			}

			return true
		}
	}
}

func replace(cur, next Interval, offset time.Duration) ([]IntervalSet, bool) {
	switch {
	// next is after
	//  c -  |###|
	//  n -        |###|
	case cur.To.Before(next.From):
		return []IntervalSet{{cur: true, i: cur}, {next: true, i: next}}, false

	// cur is after
	//  c -        |###|
	//  n -  |###|
	case cur.From.After(next.To):
		return []IntervalSet{{next: true, i: next}, {cur: true, i: cur}}, false

	// overlapping
	//  c -  |#######|
	//  n - ???#####????
	default:
		switch cur.From.Compare(next.From) {

		// cur before next
		//  c -  |#######|
		//  n -    |####????
		case -1:
			switch next.To.Compare(cur.To) {

			// after; next goes beyond cur
			//  c -  |#######|
			//  n -     |#######|
			case 1:
				return []IntervalSet{
					{cur: true, i: Interval{From: cur.From, To: next.From.Add(-offset)}},
					{next: true, i: Interval{From: next.From, To: next.To}},
				}, true

			// before; next is within cur
			//  c -  |#######|
			//  n -     |###|
			case -1:
				return []IntervalSet{
					{cur: true, i: Interval{From: cur.From, To: next.From.Add(-offset)}},
					{next: true, i: Interval{From: next.From, To: next.To}},
					{cur: true, i: Interval{From: next.To.Add(offset), To: cur.To}},
				}, true

			// case 0: matching ends
			//  c -  |#######|
			//  n -     |####|
			default:
				return []IntervalSet{
					{cur: true, i: Interval{From: cur.From, To: next.From.Add(-offset)}},
					{next: true, i: Interval{From: next.From, To: next.To}},
				}, true
			}

		// cur after next
		//  c -    |#######|
		//  n -  |####????????
		case 1:
			switch next.To.Compare(cur.To) {

			// overlap: entirety of current
			//  c -    |#######|
			//  n -  |############|
			case 1:
				return []IntervalSet{
					{next: true, i: Interval{From: next.From, To: next.To}},
				}, true

			// overlap: portion of start
			//  c -    |#######|
			//  n -  |######|
			case -1:
				return []IntervalSet{
					{next: true, i: Interval{From: next.From, To: next.To}},
					{cur: true, i: Interval{From: next.To.Add(offset), To: cur.To}},
				}, true

			// case 0: overlap: entirety of current
			//  c -    |#######|
			//  n -  |#########|
			default:
				return []IntervalSet{
					{next: true, i: Interval{From: next.From, To: next.To}},
				}, true
			}

		// cur equal to next, 0
		//  c -  |#######|
		//  n -  |####????
		default:
			switch cur.To.Compare(next.To) {

			// before
			//  c -  |#######|
			//  n -  |###########|
			case -1:
				return []IntervalSet{
					{next: true, i: Interval{From: next.From, To: next.To}},
				}, true

			// after
			//  c -  |#######|
			//  n -  |#####|
			case 1:
				return []IntervalSet{
					{next: true, i: Interval{From: next.From, To: next.To}},
					{cur: true, i: Interval{From: next.To.Add(offset), To: cur.To}},
				}, true

			// case 0: equal
			//  c -  |#######|
			//  n -  |#######|
			default:
				return []IntervalSet{
					{next: true, i: Interval{From: next.From, To: next.To}},
				}, true
			}
		}
	}
}

func split(cur, next Interval, offset time.Duration) ([]IntervalSet, bool) {
	switch {
	// next is after
	//  c -  |###|
	//  n -        |###|
	case cur.To.Before(next.From):
		return []IntervalSet{{cur: true, i: cur}, {next: true, i: next}}, false

	// cur is after
	//  c -        |###|
	//  n -  |###|
	case cur.From.After(next.To):
		return []IntervalSet{{next: true, i: next}, {cur: true, i: cur}}, false

	// overlapping
	//  c -  |#######|
	//  n - ???#####????
	default:
		switch cur.From.Compare(next.From) {

		// cur before next
		//  c -  |#######|
		//  n -    |####????
		case -1:
			switch next.To.Compare(cur.To) {

			// after; next goes beyond cur
			//  c -  |#######|
			//  n -     |#######|
			case 1:
				return []IntervalSet{
					{cur: true, i: Interval{From: cur.From, To: next.From.Add(-offset)}},
					{cur: true, next: true, i: Interval{From: next.From, To: cur.To}},
					{next: true, i: Interval{From: cur.To.Add(offset), To: next.To}},
				}, true

			// before; next is within cur
			//  c -  |#######|
			//  n -     |###|
			case -1:
				return []IntervalSet{
					{cur: true, i: Interval{From: cur.From, To: next.From.Add(-offset)}},
					{cur: true, next: true, i: Interval{From: next.From, To: next.To}},
					{cur: true, i: Interval{From: next.To.Add(offset), To: cur.To}},
				}, true

			// case 0: matching ends
			//  c -  |#######|
			//  n -     |####|
			default:
				return []IntervalSet{
					{cur: true, i: Interval{From: cur.From, To: next.From.Add(-offset)}},
					{cur: true, next: true, i: Interval{From: next.From, To: next.To}},
				}, true
			}

		// cur after next
		//  c -    |#######|
		//  n -  |####????????
		case 1:
			switch next.To.Compare(cur.To) {

			// overlap: entirety of current
			//  c -    |#######|
			//  n -  |############|
			case 1:
				return []IntervalSet{
					{next: true, i: Interval{From: next.From, To: cur.From.Add(-offset)}},
					{cur: true, next: true, i: Interval{From: cur.From, To: cur.To}},
					{next: true, i: Interval{From: cur.To.Add(offset), To: next.To}},
				}, true

			// overlap: portion of start
			//  c -    |#######|
			//  n -  |######|
			case -1:
				return []IntervalSet{
					{next: true, i: Interval{From: next.From, To: cur.From.Add(-offset)}},
					{cur: true, next: true, i: Interval{From: cur.From, To: next.To}},
					{cur: true, i: Interval{From: next.To.Add(offset), To: cur.To}},
				}, true

			// case 0: overlap: entirety of current
			//  c -    |#######|
			//  n -  |#########|
			default:
				return []IntervalSet{
					{next: true, i: Interval{From: next.From, To: cur.From.Add(-offset)}},
					{cur: true, next: true, i: Interval{From: cur.From, To: cur.To}},
				}, true
			}

		// cur equal to next, 0
		//  c -  |#######|
		//  n -  |####????
		default:
			switch cur.To.Compare(next.To) {

			// before
			//  c -  |#######|
			//  n -  |###########|
			case -1:
				return []IntervalSet{
					{cur: true, next: true, i: Interval{From: cur.From, To: cur.To}},
					{next: true, i: Interval{From: cur.To.Add(offset), To: next.To}},
				}, true

			// after
			//  c -  |#######|
			//  n -  |#####|
			case 1:
				return []IntervalSet{
					{cur: true, next: true, i: Interval{From: cur.From, To: next.To}},
					{cur: true, i: Interval{From: next.To.Add(offset), To: cur.To}},
				}, true

			// case 0: equal
			//  c -  |#######|
			//  n -  |#######|
			default:
				return []IntervalSet{
					{cur: true, next: true, i: cur},
				}, true
			}
		}
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
