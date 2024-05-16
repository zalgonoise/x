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

type mappingFunc func(cur Interval, next Interval, offset time.Duration) ([]IntervalSet, bool)

// Replace consumes the sequence Seq of Interval and data to align all From and To time.Time values for each Interval
// and coalescing the data in any intersections when required. It returns a similar but new instance of the same type of
// Seq, but with organized values.
func Replace[T any](cmpFunc func(a, b T) bool, offset time.Duration) ReducerFunc[T] {
	return func(seq SeqKV[Interval, T]) (reduced SeqKV[Interval, T]) {
		cache := make([]DataInterval[T], 0, minAlloc)

		if !seq(func(interval Interval, t T) bool {
			cache = resolveConflicts[T](
				interval, t, cache,
				replace,
				cmpFunc,
				func(a, b T) T { return b },
				offset,
			)

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

func splitConflicts[T any](conflicts []DataInterval[T], interval Interval, value T) []Interval {
	var (
		head = interval.From
		tail = interval.To
		lim  time.Time
	)

	splitIntervals := make([]Interval, 0, len(conflicts))

	for i := len(conflicts) - 1; i >= 0; i-- {
		if i == 0 {
			lim = head
		} else {
			lim = getNearestLimit(interval, conflicts[i-1].Interval)
		}

		splitIntervals = append(splitIntervals, Interval{From: lim, To: tail})

		tail = lim
	}

	slices.Reverse(splitIntervals)

	return splitIntervals
}

func getNearestLimit(interval, conflict Interval) time.Time {
	switch interval.From.Compare(conflict.To) {
	case 1:
		return interval.To
	default:
		return conflict.To
	}
}

func resolveConflicts[T any](
	interval Interval, value T, cache []DataInterval[T],
	mapFunc mappingFunc, cmpFunc func(a, b T) bool, mergeFunc func(a, b T) T,
	offset time.Duration,
) []DataInterval[T] {
	cache = mergeCache(cache, cmpFunc, offset)

	conflicts, indices := findConflicts(cache, interval)

	switch len(conflicts) {
	case 0:
		cache = append(cache, DataInterval[T]{
			Data:     value,
			Interval: interval,
		})

		return cache
	case 1:
		return resolveConflict[T](cache, indices[0], conflicts[0], interval, value, offset, mapFunc, cmpFunc, mergeFunc)

	default:
		conflictSet := splitConflicts(conflicts, interval, value)
		tempCache := make([]DataInterval[T], 0, len(conflicts)*3)

		for i := len(conflicts) - 1; i >= 0; i-- {
			sets, overlaps := mapFunc(conflicts[i].Interval, conflictSet[i], offset)

			if !overlaps {
				continue
			}

			for idx := range sets {
				switch {
				case sets[idx].cur && !sets[idx].next:
					tempCache = append(tempCache, DataInterval[T]{
						Data:     conflicts[i].Data,
						Interval: sets[idx].i,
					})
				case !sets[idx].cur && sets[idx].next:
					tempCache = append(tempCache, DataInterval[T]{
						Data:     value,
						Interval: sets[idx].i,
					})
				default:
					tempCache = append(tempCache, DataInterval[T]{
						Data:     mergeFunc(conflicts[i].Data, value),
						Interval: sets[idx].i,
					})
				}
			}
		}

		for i := len(conflicts) - 1; i >= 0; i-- {
			cache = slices.Delete(cache, indices[i], indices[i]+1)
		}

		cache = append(cache, tempCache...)

		slices.SortFunc(cache, func(a, b DataInterval[T]) int {
			return a.Interval.From.Compare(b.Interval.From)
		})

		return mergeCache(cache, cmpFunc, offset)
	}
}

func resolveConflict[T any](
	cache []DataInterval[T], index int, conflict DataInterval[T],
	interval Interval, value T, offset time.Duration,
	mapFunc mappingFunc, cmpFunc func(a, b T) bool, mergeFunc func(a, b T) T,
) (resolved []DataInterval[T]) {
	cache = slices.Delete(cache, index, index+1)

	sets, overlaps := mapFunc(conflict.Interval, interval, offset)

	if !overlaps {
		return cache
	}

	for idx := range sets {
		switch {
		case sets[idx].cur && !sets[idx].next:
			cache = append(cache, DataInterval[T]{
				Data:     conflict.Data,
				Interval: sets[idx].i,
			})
		case !sets[idx].cur && sets[idx].next:
			cache = append(cache, DataInterval[T]{
				Data:     value,
				Interval: sets[idx].i,
			})
		default:
			cache = append(cache, DataInterval[T]{
				Data:     mergeFunc(conflict.Data, value),
				Interval: sets[idx].i,
			})
		}
	}

	slices.SortFunc(cache, func(a, b DataInterval[T]) int {
		return a.Interval.From.Compare(b.Interval.From)
	})

	return mergeCache(cache, cmpFunc, offset)
}

// Flatten consumes the sequence Seq of Interval and data to align all From and To time.Time values for each Interval
// and coalescing the data in any intersections when required. It returns a similar but new instance of the same type of
// Seq, but with organized values.
func Flatten[T any](cmpFunc func(a, b T) bool, mergeFunc func(a, b T) T, offset time.Duration) ReducerFunc[T] {
	return func(seq SeqKV[Interval, T]) (reduced SeqKV[Interval, T]) {
		cache := make([]DataInterval[T], 0, minAlloc)

		if !seq(func(interval Interval, t T) bool {
			cache = resolveConflicts(interval, t, cache, split, cmpFunc, mergeFunc, offset)

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

func mergeCache[T any](cache []DataInterval[T], cmp func(a, b T) bool, offset time.Duration) []DataInterval[T] {
	if len(cache) <= 1 {
		return cache
	}

	// iterate backwards merging equal ranges
	//
	// |#### D #####|
	//              |## E ##|
	//    |# D #|   |##### E #####|
	//                            |## E ##|
	//
	// |#### D #####|######### E #########|
	for i := len(cache) - 1; i > 0; i-- {
		switch {
		// cur starts where prev ends; cur eq prev
		case cache[i-1].Interval.To.Sub(cache[i].Interval.From) <= offset &&
			cmp(cache[i-1].Data, cache[i].Data):
			cache[i-1].Interval.To = cache[i].Interval.To
			cache = slices.Delete(cache, i, i+1)

		// prev is within cur; cur eq prev
		case cache[i-1].Interval.To.Before(cache[i].Interval.To) &&
			cache[i-1].Interval.From.After(cache[i].Interval.From) &&
			cmp(cache[i-1].Data, cache[i].Data):
			cache = slices.Delete(cache, i-1, i)

		// cur is within prev; cur eq prev
		case cache[i].Interval.To.Before(cache[i-1].Interval.To) &&
			cache[i].Interval.From.After(cache[i-1].Interval.From) &&
			cmp(cache[i-1].Data, cache[i].Data):
			cache = slices.Delete(cache, i, i+1)

		default:
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
