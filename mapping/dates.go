package mapping

import (
	"errors"
	"maps"
	"slices"
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

type Timeframe[K comparable, T any] struct {
	noOverlap bool

	Index *Index[Interval, map[K]T]
}

type Interval struct {
	From time.Time
	To   time.Time
}

type IntervalSet struct {
	cur  bool
	next bool
	i    Interval
}

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

	if _, ok := t.Index.Get(i); ok {
		t.Index.Set(i, func(old map[K]T) (map[K]T, bool) {
			return coalesce(old, values), true
		})

		return nil
	}

	t.Index.Set(i, func(old map[K]T) (map[K]T, bool) {
		return values, true
	})

	return nil
}

func (t *Timeframe[K, T]) Organize() (*Timeframe[K, T], error) {
	cache := make([]DataInterval[K, T], 0, len(t.Index.values)*2)
	keys := make([]Interval, len(t.Index.Keys))
	copy(keys, t.Index.Keys)

	slices.SortFunc(keys, func(a, b Interval) int {
		return a.From.Compare(b.From)
	})

	for len(keys) > 0 {
		key := keys[0]
		keys = keys[1:]

		if len(cache) == 0 {
			cache = append(cache, DataInterval[K, T]{
				Data:     t.Index.values[key],
				Interval: key,
			})

			continue
		}

		conflicts := findConflicts(cache, key)

		if len(conflicts) == 0 {
			cache = append(cache, DataInterval[K, T]{
				Data:     t.Index.values[key],
				Interval: key,
			})

			continue
		}

		for i := range conflicts {
			sets, overlaps, err := split(conflicts[i].Interval, key)
			if err != nil {
				return nil, err
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
						Data:     t.Index.values[key],
						Interval: sets[idx].i,
					})
				default:
					dataCopy := maps.Clone(t.Index.values[key])
					coalesce(dataCopy, conflicts[i].Data)

					cache = append(cache, DataInterval[K, T]{
						Data:     dataCopy,
						Interval: sets[idx].i,
					})
				}
			}
		}
	}

	tf := NewTimeframe[K, T]()

	for i := range cache {
		if err := tf.Add(cache[i].Interval, cache[i].Data); err != nil {
			return nil, err
		}
	}

	return tf, nil
}

// All returns an iterator over the values in the Timeframe,
// through the indexed Interval values ordered by their From time.Time value.
func (t *Timeframe[K, T]) All() SeqKV[Interval, map[K]T] {
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

func Organize[K comparable, T any](seq SeqKV[Interval, map[K]T]) (tf *Timeframe[K, T], err error) {
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

	tf = NewTimeframe[K, T]()

	for i := range cache {
		if err = tf.Add(cache[i].Interval, cache[i].Data); err != nil {
			return nil, err
		}
	}

	return tf, nil
}

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

// Append iterates through the input SeqKV and adds all intervals and respective values
// to the Timeframe t.
func (t *Timeframe[K, T]) Append(seq SeqKV[Interval, map[K]T]) (err error) {
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
// by extracting a SeqKV of the same items from tf using Timeframe.All, and adding them into
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
