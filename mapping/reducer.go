package mapping

import (
	"iter"
	"slices"
	"time"

	"github.com/zalgonoise/cfg"
)

// ReducerFunc transforms an input sequence of (Interval, T) pairs into an
// organized sequence of (Interval, T) pairs according to a conflict-resolution
// strategy.
type ReducerFunc[T any] func(iter.Seq2[Interval, T]) iter.Seq2[Interval, T]

// NewReducer creates a reusable [ReducerFunc] configured with opts.
func NewReducer[T any](opts ...cfg.Option[Config[T]]) ReducerFunc[T] {
	config := configure(Config[T]{}, opts...)

	return func(seq iter.Seq2[Interval, T]) iter.Seq2[Interval, T] {
		return organizeSeqWithConfig(seq, config)
	}
}

// Organize processes a slice of [DataInterval] items according to the strategy
// and options specified in opts, returning a newly allocated slice of organized
// [DataInterval] items ordered chronologically.
//
// By default, if no strategy option is provided, [WithReplace] is used.
func Organize[T any](data []DataInterval[T], opts ...cfg.Option[Config[T]]) []DataInterval[T] {
	config := configure(Config[T]{}, opts...)

	return organizeSlice(data, config)
}

// OrganizeSeq consumes a sequence of (Interval, T) pairs and returns a new
// [iter.Seq2] yielding the organized intervals and values according to opts.
//
// By default, if no strategy option is provided, [WithReplace] is used.
func OrganizeSeq[T any](seq iter.Seq2[Interval, T], opts ...cfg.Option[Config[T]]) iter.Seq2[Interval, T] {
	config := configure(Config[T]{}, opts...)

	return organizeSeqWithConfig(seq, config)
}

func organizeSeqWithConfig[T any](seq iter.Seq2[Interval, T], config Config[T]) iter.Seq2[Interval, T] {
	if seq == nil {
		return func(func(Interval, T) bool) {}
	}

	if config.strategy == strategyAppend {
		items := make([]DataInterval[T], 0, minAlloc)
		for interval, val := range seq {
			if !isValidInterval(interval, config.offset) {
				continue
			}

			items = append(items, DataInterval[T]{
				Interval: interval,
				Data:     val,
			})
		}

		slices.SortStableFunc(items, func(a, b DataInterval[T]) int {
			return a.Interval.Compare(b.Interval)
		})

		return AsSeq(items)
	}

	var (
		cache   = make([]DataInterval[T], 0, minAlloc)
		scratch []DataInterval[T]
		seeded  = false
	)

	for interval, val := range seq {
		if !isValidInterval(interval, config.offset) {
			continue
		}

		cache, scratch = resolveStep(cache, scratch, interval, val, config, seeded)
		seeded = true
	}

	return AsSeq(cache)
}

func organizeSlice[T any](data []DataInterval[T], config Config[T]) []DataInterval[T] {
	if len(data) == 0 {
		return []DataInterval[T]{}
	}

	if config.strategy == strategyAppend {
		out := make([]DataInterval[T], 0, len(data))
		for i := range data {
			if !isValidInterval(data[i].Interval, config.offset) {
				continue
			}

			out = append(out, data[i])
		}

		slices.SortStableFunc(out, func(a, b DataInterval[T]) int {
			return a.Interval.Compare(b.Interval)
		})

		return out
	}

	var (
		cache   = make([]DataInterval[T], 0, max(len(data), minAlloc))
		scratch []DataInterval[T]
		seeded  = false
	)

	for i := range data {
		if !isValidInterval(data[i].Interval, config.offset) {
			continue
		}

		cache, scratch = resolveStep(cache, scratch, data[i].Interval, data[i].Data, config, seeded)
		seeded = true
	}

	return cache
}

func resolveStep[T any](
	cache []DataInterval[T],
	scratch []DataInterval[T],
	next Interval,
	value T,
	config Config[T],
	seeded bool,
) ([]DataInterval[T], []DataInterval[T]) {
	if config.strategy == strategyAppend {
		idx, _ := slices.BinarySearchFunc(cache, next, func(elem DataInterval[T], target Interval) int {
			if cmp := elem.Interval.Compare(target); cmp != 0 {
				return cmp
			}

			return -1
		})

		return slices.Insert(cache, idx, DataInterval[T]{
			Interval: next,
			Data:     value,
		}), scratch
	}

	if !seeded {
		return append(cache, DataInterval[T]{
			Interval: next,
			Data:     value,
		}), scratch
	}

	if len(cache) == 0 {
		if config.strategy == strategyIntersect {
			return cache, scratch
		}

		return append(cache, DataInterval[T]{
			Interval: next,
			Data:     value,
		}), scratch
	}

	start, end := findOverlapRange(cache, next, config.offset)

	if start == end {
		if config.strategy == strategyIntersect {
			clear(cache)
			return cache[:0], scratch
		}

		return insertDisjoint(cache, start, next, value, config.cmpFunc, config.offset), scratch
	}

	var (
		resolved    = scratch[:0]
		isIntersect = config.strategy == strategyIntersect
		offset      = config.offset
		cursor      = next.From
	)

	if cap(resolved) < (end-start)*3 {
		resolved = make([]DataInterval[T], 0, max((end-start)*3, 16))
	}

	for i := start; i < end; i++ {
		c := cache[i]

		// 1. Non-overlapping prefix of existing interval c before next.From.
		if !isIntersect && c.Interval.From.Before(next.From) {
			prefixEnd := next.From.Add(-offset)
			if validSubInterval(c.Interval.From, prefixEnd, offset) {
				resolved = append(resolved, DataInterval[T]{
					Interval: Interval{From: c.Interval.From, To: prefixEnd},
					Data:     c.Data,
				})
			}
		}

		// 2. Non-overlapping gap of incoming interval next before c.Interval.From.
		if cursor.Before(c.Interval.From) {
			if !isIntersect {
				gapEnd := c.Interval.From.Add(-offset)
				if validSubInterval(cursor, gapEnd, offset) {
					resolved = append(resolved, DataInterval[T]{
						Interval: Interval{From: cursor, To: gapEnd},
						Data:     value,
					})
				}
			}

			cursor = c.Interval.From
		}

		// 3. Intersecting segment between c and next.
		overlapStart := maxTime(cursor, c.Interval.From)
		overlapEnd := minTime(c.Interval.To, next.To)

		if validSubInterval(overlapStart, overlapEnd, offset) {
			resolved = append(resolved, DataInterval[T]{
				Interval: Interval{From: overlapStart, To: overlapEnd},
				Data:     config.mergeFunc(c.Data, value),
			})
		}

		cursor = overlapEnd.Add(offset)

		// 4. Non-overlapping suffix of existing interval c after next.To.
		if !isIntersect && c.Interval.To.After(next.To) {
			suffixStart := next.To.Add(offset)
			if validSubInterval(suffixStart, c.Interval.To, offset) {
				resolved = append(resolved, DataInterval[T]{
					Interval: Interval{From: suffixStart, To: c.Interval.To},
					Data:     c.Data,
				})
			}
		}
	}

	// 5. Remaining non-overlapping tail of incoming interval next after last conflict.
	if !isIntersect && validSubInterval(cursor, next.To, offset) {
		resolved = append(resolved, DataInterval[T]{
			Interval: Interval{From: cursor, To: next.To},
			Data:     value,
		})
	}

	// Coalesce adjacent equal segments within resolved before splicing into cache.
	resolved = coalesce(resolved, config.cmpFunc, offset)

	if isIntersect {
		clear(cache)
		cache = append(cache[:0], resolved...)

		return cache, resolved[:0]
	}

	// Absorb left and right neighbors in cache if they touch resolved and carry equal Data.
	if len(resolved) > 0 && config.cmpFunc != nil {
		if start > 0 &&
			resolved[0].Interval.From.Sub(cache[start-1].Interval.To) <= offset &&
			config.cmpFunc(cache[start-1].Data, resolved[0].Data) {
			resolved[0].Interval.From = cache[start-1].Interval.From
			resolved[0].Data = cache[start-1].Data
			start--
		}

		last := len(resolved) - 1
		if end < len(cache) &&
			cache[end].Interval.From.Sub(resolved[last].Interval.To) <= offset &&
			config.cmpFunc(resolved[last].Data, cache[end].Data) {
			resolved[last].Interval.To = cache[end].Interval.To
			end++
		}
	}

	cache = slices.Replace(cache, start, end, resolved...)

	return cache, resolved[:0]
}

func insertDisjoint[T any](
	cache []DataInterval[T],
	idx int,
	next Interval,
	value T,
	cmpFunc CompareFunc[T],
	offset time.Duration,
) []DataInterval[T] {
	if cmpFunc != nil {
		mergedLeft := idx > 0 &&
			next.From.Sub(cache[idx-1].Interval.To) <= offset &&
			cmpFunc(cache[idx-1].Data, value)

		mergedRight := idx < len(cache) &&
			cache[idx].Interval.From.Sub(next.To) <= offset &&
			cmpFunc(value, cache[idx].Data)

		switch {
		case mergedLeft && mergedRight:
			cache[idx-1].Interval.To = cache[idx].Interval.To
			return slices.Delete(cache, idx, idx+1)
		case mergedLeft:
			cache[idx-1].Interval.To = next.To
			return cache
		case mergedRight:
			cache[idx].Interval.From = next.From
			cache[idx].Data = value
			return cache
		}
	}

	return slices.Insert(cache, idx, DataInterval[T]{
		Interval: next,
		Data:     value,
	})
}

// findOverlapRange returns the half-open index range [start, end) of intervals
// in the sorted, disjoint cache that overlap with cur.
func findOverlapRange[T any](cache []DataInterval[T], cur Interval, offset time.Duration) (start, end int) {
	if offset > 0 {
		// With a positive discrete offset, bounds [From, To] are inclusive:
		// intervals overlap when c.To >= cur.From and c.From <= cur.To.
		start, _ = slices.BinarySearchFunc(cache, cur.From, func(elem DataInterval[T], target time.Time) int {
			if elem.Interval.To.Before(target) {
				return -1
			}

			return 1
		})

		relEnd, _ := slices.BinarySearchFunc(cache[start:], cur.To, func(elem DataInterval[T], target time.Time) int {
			if !elem.Interval.From.After(target) {
				return -1
			}

			return 1
		})

		return start, start + relEnd
	}

	// With offset == 0, bounds [From, To) are half-open:
	// intervals overlap when c.To > cur.From and c.From < cur.To.
	start, _ = slices.BinarySearchFunc(cache, cur.From, func(elem DataInterval[T], target time.Time) int {
		if elem.Interval.To.Compare(target) <= 0 {
			return -1
		}

		return 1
	})

	relEnd, _ := slices.BinarySearchFunc(cache[start:], cur.To, func(elem DataInterval[T], target time.Time) int {
		if elem.Interval.From.Compare(target) < 0 {
			return -1
		}

		return 1
	})

	return start, start + relEnd
}

// coalesce merges adjacent or contained intervals in a chronologically sorted
// slice when their gap is at most offset and cmpFunc reports their Data values
// are equal.
func coalesce[T any](cache []DataInterval[T], cmpFunc CompareFunc[T], offset time.Duration) []DataInterval[T] {
	if len(cache) <= 1 || cmpFunc == nil {
		return cache
	}

	write := 0

	for read := 1; read < len(cache); read++ {
		if cache[read].Interval.From.Sub(cache[write].Interval.To) <= offset &&
			cmpFunc(cache[write].Data, cache[read].Data) {
			if cache[read].Interval.To.After(cache[write].Interval.To) {
				cache[write].Interval.To = cache[read].Interval.To
			}

			continue
		}

		write++
		if write != read {
			cache[write] = cache[read]
		}
	}

	clear(cache[write+1:])

	return cache[:write+1]
}

func isValidInterval(i Interval, offset time.Duration) bool {
	if !i.IsValid() {
		return false
	}

	if offset == 0 && i.From.Equal(i.To) {
		return false
	}

	return true
}

func validSubInterval(from, to time.Time, offset time.Duration) bool {
	if offset == 0 {
		return from.Before(to)
	}

	return !to.Before(from)
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}

	return b
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}

	return b
}
