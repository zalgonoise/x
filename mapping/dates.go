package mapping

import (
	"cmp"
	"errors"
	"time"
)

const defaultMaxGap = time.Minute

var (
	errZeroOrNegativeDur     = errors.New("interval cannot end at the same time or before start")
	errOverlappingTimeframes = errors.New("overlapping timeframes")
	errGapBetweenTimeframes  = errors.New("gap between timeframes")
	errOrganizeFailed        = errors.New("failed to organize timeframe")
	errIntervalNotFound      = errors.New("interval not found")
)

type Interval struct {
	From time.Time
	// TODO: benchmark this approach; if another time.Time for a To value isn't preferable
	To time.Time
}

type KV[K comparable, T any] struct {
	// TODO: should *Timeframe[K, T].Add(Interval, []KV) accept an *Index[K, T] instead?
	Valid bool

	Key   K
	Value T
}

type Timeframe[K comparable, T any] struct {
	noOverlap bool
	maxGap    time.Duration

	Index *Index[Interval, *Index[K, T]]
}

type IntervalSet struct {
	cur  bool
	next bool
	i    Interval
}

func split(cur, next Interval) []IntervalSet {
	switch {
	case cur.To.Before(next.From):
		return []IntervalSet{{cur: true, i: cur}, {next: true, i: next}}
	case cur.From.After(next.To):
		return []IntervalSet{{next: true, i: next}, {cur: true, i: cur}}
	case cur.From.Equal(next.From):
		switch cur.To.Compare(next.To) {
		case -1: // before
			return []IntervalSet{
				{cur: true, next: true, i: Interval{From: cur.From, To: cur.To}},
				{next: true, i: Interval{From: cur.To, To: next.To}},
			}
		case 1: // after
			return []IntervalSet{
				{cur: true, next: true, i: Interval{From: cur.From, To: next.To}},
				{cur: true, i: Interval{From: next.To, To: cur.To}},
			}
		case 0: // equal
			return []IntervalSet{
				{cur: true, next: true, i: cur},
			}
		}

	// overlap: portion of end
	case next.From.After(cur.From):
		switch next.To.Compare(cur.To) {
		case 1: // after; next goes beyond cur
			return []IntervalSet{
				{cur: true, i: Interval{From: cur.From, To: next.From}},
				{cur: true, next: true, i: Interval{From: next.From, To: cur.To}},
				{next: true, i: Interval{From: cur.To, To: next.To}},
			}
		case -1: // before; next is within cur
			return []IntervalSet{
				{cur: true, i: Interval{From: cur.From, To: next.From}},
				{cur: true, next: true, i: Interval{From: next.From, To: next.To}},
				{next: true, i: Interval{From: next.To, To: cur.To}},
			}
		case 0:
			return []IntervalSet{
				{cur: true, i: Interval{From: cur.From, To: next.From}},
				{cur: true, next: true, i: Interval{From: next.From, To: next.To}},
			}
		}

	// overlap: portion of start
	case next.From.Before(cur.From):
		switch next.To.Compare(cur.To) {
		case 1:
			return []IntervalSet{
				{next: true, i: Interval{From: next.From, To: cur.From}},
				{cur: true, next: true, i: Interval{From: cur.From, To: cur.To}},
				{next: true, i: Interval{From: cur.To, To: next.To}},
			}
		case -1:
			return []IntervalSet{
				{next: true, i: Interval{From: next.From, To: cur.From}},
				{cur: true, next: true, i: Interval{From: cur.From, To: next.To}},
				{cur: true, i: Interval{From: next.To, To: cur.To}},
			}
		case 0:
			return []IntervalSet{
				{next: true, i: Interval{From: next.From, To: cur.From}},
				{cur: true, next: true, i: Interval{From: cur.From, To: cur.To}},
			}
		}
	}

	return nil
}

//func (t *Timeframe[K, T]) organize(idx int, i Interval, values []KV[K, T]) error {
//	// format input
//	keys := make([]K, 0, len(values))
//	vals := make(map[K]T, len(values))
//
//	for vi := range values {
//		keys = append(keys, values[vi].Key)
//		vals[values[vi].Key] = values[vi].Value
//	}
//
//	// retrieve old entry
//	storedInterval := t.Index.Keys[idx]
//	storedValues, ok := t.Index.values[storedInterval]
//	if !ok {
//		return fmt.Errorf("%w: from: %s ; to: %s",
//			errIntervalNotFound,
//			storedInterval.From.String(), storedInterval.To.String(),
//		)
//	}
//
//	// remove old entry
//	t.Index.Keys = append(t.Index.Keys[:idx], t.Index.Keys[idx+1:]...)
//	delete(t.Index.values, storedInterval)
//
//	// modify entries
//	switch {
//	// overlap: portion of end
//	case storedInterval.From.Before(i.From) && storedInterval.To.After(i.From):
//		t.Index.Set(Interval{
//			From: storedInterval.From,
//			To:   i.From,
//		}, func(old *Index[K, T]) (*Index[K, T], bool) {
//			return storedValues, true
//		})
//
//		switch {
//		case storedInterval.To.After(i.To):
//			t.Index.Set(Interval{
//				From: i.From,
//				To:   i.To,
//			}, func(old *Index[K, T]) (*Index[K, T], bool) {
//				return &Index[K, T]{
//					Keys:   slices.Concat(storedValues.Keys, keys),
//					values: coalesce(vals, storedValues.values),
//				}, true
//			})
//
//			t.Index.Set(Interval{
//				From: i.To,
//				To:   storedInterval.To,
//			}, func(old *Index[K, T]) (*Index[K, T], bool) {
//				return storedValues, true
//			})
//		default:
//		}
//
//	// overlap: portion of start
//	case storedInterval.From.Before(i.To) && storedInterval.To.After(i.To):
//
//	}
//
//	// store modified entries
//
//	return nil
//}

func (t *Timeframe[K, T]) Add(i Interval, values []KV[K, T]) error {
	// check overlaps
	intervals := t.Index.Keys

	for idx := range intervals {
		switch {
		// ends before this one starts
		case intervals[idx].To.Before(i.From):
			continue
		// starts before this one ends
		case intervals[idx].From.After(i.To):
			continue
		default:
			//return t.organize(idx, i, values)
		}
	}

	// insert new data
	t.Index.Set(i, func(old *Index[K, T]) (*Index[K, T], bool) {
		var set bool

		if old == nil {
			old = &Index[K, T]{
				Keys:   make([]K, 0, len(values)),
				values: make(map[K]T, len(values)),
			}
		}

		for idx := range values {
			old.Set(values[idx].Key, func(old T) (T, bool) {
				return values[idx].Value, true
			})

			set = true
		}

		return old, set
	})

	return nil
}

//func (t *Timeframe[K, T]) Add(i Interval, values []KV[K, T]) error {
//	if i.To.Before(i.From) {
//		return errZeroOrNegativeDur
//	}
//
//	t.Index.Set(i, func(old *Index[K, T]) (*Index[K, T], bool) {
//		if old == nil {
//			old = &Index[K, T]{
//				Keys:   make([]K, 0, len(values)),
//				values: make(map[K]T, len(values)),
//			}
//		}
//
//		var set bool
//
//		for idx := range values {
//			if values[idx].Valid {
//				set = true
//
//				old.Set(values[idx].Key, func(old T) (T, bool) {
//					return values[idx].Value, true
//				})
//			}
//		}
//
//		return old, set
//	})
//
//	return nil
//}

type cache[K comparable, T any] struct {
	interval Interval
	value    *Index[K, T]
}

// ref: https://github.com/golang/go/issues/61899

type Seq[T, K any] func(yield func(T, K) bool) bool

// All returns an iterator over the values in the slice,
// starting with s[0].
func (t *Timeframe[K, T]) All() Seq[Interval, *Index[K, T]] {
	return func(yield func(Interval, *Index[K, T]) bool) bool {
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

//func (t *Timeframe[K, T]) Organize(seq Seq[Interval, *Index[K, T]]) (*Timeframe[K, T], error) {
//	// TODO: tidy up this method; break it down however possible
//
//	var (
//		c   *cache[K, T]
//		err error
//
//		tf = &Timeframe[K, T]{
//			noOverlap: t.noOverlap,
//			maxGap:    t.maxGap,
//			Index: &Index[Interval, *Index[K, T]]{
//				Keys:   make([]Interval, 0, len(t.Index.Keys)),
//				zero:   t.Index.zero,
//				cmp:    t.Index.cmp,
//				values: make(map[Interval]*Index[K, T], len(t.Index.values)),
//			},
//		}
//	)
//
//	if !seq(func(interval Interval, i *Index[K, T]) bool {
//		if c == nil {
//			c = &cache[K, T]{
//				interval: interval,
//				value:    i,
//			}
//
//			return true
//		}
//
//		prevStart := c.interval.From.UnixNano()
//		curStart := interval.From.UnixNano()
//		prevEnd := prevStart + int64(c.interval.Dur)
//
//		switch {
//		case prevStart == curStart:
//			// w/ overlap
//			if t.noOverlap {
//				err = errOverlappingTimeframes
//
//				return false
//			}
//
//			curEnd := curStart + int64(interval.Dur)
//
//			switch {
//			case curEnd == prevEnd:
//				c.value.values = coalesce(c.value.values, i.values)
//
//				return true
//			case prevEnd < curEnd:
//				firstHalf := c.interval.Dur
//				secHalf := interval.Dur - firstHalf
//
//				if !tf.Index.Set(Interval{
//					From: c.interval.From,
//					Dur:  firstHalf - 1,
//				}, func(old *Index[K, T]) (*Index[K, T], bool) {
//					if old == nil {
//						old = &Index[K, T]{
//							values: coalesce(c.value.values, i.values),
//						}
//
//						old.Keys = make([]K, 0, len(old.values))
//
//						for key := range old.values {
//							old.Keys = append(old.Keys, key)
//						}
//
//						return old, true
//					}
//
//					old.values = coalesce(c.value.values, i.values)
//
//					return old, true
//				}) {
//					return false
//				}
//
//				c.interval = Interval{
//					From: interval.From.Add(firstHalf),
//					Dur:  secHalf,
//				}
//				c.value = i
//
//				return true
//
//			case curEnd < prevEnd:
//				firstHalf := interval.Dur
//				secHalf := c.interval.Dur - firstHalf
//
//				if !tf.Index.Set(Interval{
//					From: c.interval.From,
//					Dur:  firstHalf - 1,
//				}, func(old *Index[K, T]) (*Index[K, T], bool) {
//					if old == nil {
//						old = &Index[K, T]{
//							values: coalesce(c.value.values, i.values),
//						}
//
//						old.Keys = make([]K, 0, len(old.values))
//
//						for key := range old.values {
//							old.Keys = append(old.Keys, key)
//						}
//
//						return old, true
//					}
//
//					old.values = coalesce(c.value.values, i.values)
//
//					return old, true
//				}) {
//					return false
//				}
//
//				c.interval = Interval{
//					From: interval.From.Add(firstHalf),
//					Dur:  interval.Dur - secHalf,
//				}
//
//				return true
//			}
//
//		case prevStart < curStart:
//			curEnd := curStart + int64(interval.Dur)
//
//			switch {
//			case prevEnd > curEnd:
//				// w/ overlap
//				if t.noOverlap {
//					err = errOverlappingTimeframes
//
//					return false
//				}
//
//				firstHalf := time.Duration(prevStart - curStart)
//
//				if !tf.Index.Set(Interval{
//					From: c.interval.From,
//					Dur:  firstHalf - 1,
//				}, func(old *Index[K, T]) (*Index[K, T], bool) {
//					return c.value, true
//				}) {
//					return false
//				}
//
//				if !tf.Index.Set(Interval{
//					From: interval.From,
//					Dur:  interval.Dur - 1,
//				}, func(old *Index[K, T]) (*Index[K, T], bool) {
//					if old == nil {
//						old = &Index[K, T]{
//							values: coalesce(i.values, c.value.values),
//						}
//
//						old.Keys = make([]K, 0, len(old.values))
//
//						for key := range old.values {
//							old.Keys = append(old.Keys, key)
//						}
//
//						return old, true
//					}
//
//					old.values = coalesce(i.values, c.value.values)
//
//					return old, true
//				}) {
//					return false
//				}
//
//				c.interval.From = c.interval.From.Add(firstHalf + interval.Dur)
//
//			case prevEnd > curStart:
//				// w/ overlap
//				if t.noOverlap {
//					err = errOverlappingTimeframes
//
//					return false
//				}
//
//				firstHalf := time.Duration(curStart - prevStart)
//				secHalf := c.interval.Dur - firstHalf
//
//				if !tf.Index.Set(Interval{
//					From: c.interval.From,
//					Dur:  firstHalf - 1,
//				}, func(old *Index[K, T]) (*Index[K, T], bool) {
//					return c.value, true
//				}) {
//					return false
//				}
//
//				if !tf.Index.Set(Interval{
//					From: interval.From,
//					Dur:  secHalf,
//				}, func(old *Index[K, T]) (*Index[K, T], bool) {
//					if old == nil {
//						old = &Index[K, T]{
//							values: coalesce(c.value.values, i.values),
//						}
//
//						old.Keys = make([]K, 0, len(old.values))
//
//						for key := range old.values {
//							old.Keys = append(old.Keys, key)
//						}
//
//						return old, true
//					}
//
//					old.values = coalesce(c.value.values, i.values)
//
//					return old, true
//				}) {
//					return false
//				}
//
//				c.interval = Interval{
//					From: interval.From.Add(secHalf),
//					Dur:  interval.Dur - secHalf,
//				}
//				c.value = i
//
//				return true
//
//			case prevEnd+int64(t.maxGap) < curStart:
//				// w/ gap
//				err = fmt.Errorf("%w: %s", errGapBetweenTimeframes, time.Duration(curStart-prevEnd).String())
//
//				return false
//			default:
//				ok := tf.Index.Set(c.interval, func(old *Index[K, T]) (*Index[K, T], bool) {
//					return c.value, true
//				})
//
//				c.interval = interval
//				c.value = i
//
//				if !ok {
//					return false
//				}
//			}
//		}
//
//		return true
//	}) {
//		if err != nil {
//			return nil, err
//		}
//
//		return nil, errOrganizeFailed
//	}
//
//	if c != nil {
//		if !tf.Index.Set(c.interval, func(old *Index[K, T]) (*Index[K, T], bool) {
//			return c.value, true
//		}) {
//			return nil, errOrganizeFailed
//		}
//	}
//
//	return tf, nil
//}

func NewTimeframe[K comparable, T any]() *Timeframe[K, T] {
	return &Timeframe[K, T]{
		maxGap: defaultMaxGap,
		Index: NewIndex[Interval, *Index[K, T]](
			make(map[Interval]*Index[K, T]),
			WithIndex[*Index[K, T]](func(a, b Interval) int {
				return cmp.Compare(a.From.Unix(), b.From.Unix())
			}),
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
