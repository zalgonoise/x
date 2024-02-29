package mapping

import (
	"maps"
	"slices"
)

const minAlloc = 64

type TimeframeV2[K comparable, T any] struct {
	keys   []Interval
	values map[Interval]map[K]T
}

type DataInterval[K comparable, T any] struct {
	Data     map[K]T
	Interval Interval
}

func (t *TimeframeV2[K, T]) Add(i Interval, values map[K]T) error {
	if val, ok := t.values[i]; ok {
		t.values[i] = coalesce(val, values)

		return nil
	}

	t.keys = append(t.keys, i)
	t.values[i] = values

	return nil
}

func (t *TimeframeV2[K, T]) Organize() (*TimeframeV2[K, T], error) {
	cache := make([]DataInterval[K, T], 0, len(t.keys)*2)
	final := make([]DataInterval[K, T], 0, len(t.keys)*2)
	keys := make([]Interval, len(t.keys))
	copy(keys, t.keys)

	slices.SortFunc(keys, func(a, b Interval) int {
		return a.From.Compare(b.From)
	})

	for len(keys) > 0 {
		if len(cache) == 0 {
			cache = append(cache, DataInterval[K, T]{
				Data:     t.values[keys[0]],
				Interval: keys[0],
			})

			keys = keys[1:]

			continue
		}

		key := keys[0]
		conflicts := findConflicts(cache, key)

		for i := range conflicts {
			sets, overlaps, err := split(conflicts[i].Interval, key)
			if err != nil {
				return nil, err
			}

			if !overlaps {
				continue
			}

			cache = append(cache[:i], cache[i+1:]...)

			for idx := range sets {
				switch {
				case sets[idx].cur && !sets[idx].next:
					final = append(final, DataInterval[K, T]{
						Data:     conflicts[i].Data,
						Interval: sets[idx].i,
					})
				case !sets[idx].cur && sets[idx].next:
					final = append(final, DataInterval[K, T]{
						Data:     t.values[key],
						Interval: sets[idx].i,
					})
				default:
					dataCopy := maps.Clone(t.values[key])
					coalesce(dataCopy, conflicts[i].Data)

					final = append(final, DataInterval[K, T]{
						Data:     dataCopy,
						Interval: sets[idx].i,
					})
				}
			}
		}
	}

	tf := NewTimeframeV2[K, T]()

	for i := range final {
		if err := tf.Add(final[i].Interval, final[i].Data); err != nil {
			return nil, err
		}
	}

	for i := range cache {
		if err := tf.Add(cache[i].Interval, cache[i].Data); err != nil {
			return nil, err
		}
	}

	return tf, nil
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
				{next: true, i: Interval{From: next.To, To: cur.To}},
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

	return nil, false, errTimeSplitFailed
}

func NewTimeframeV2[K comparable, T any]() *TimeframeV2[K, T] {
	return &TimeframeV2[K, T]{
		keys:   make([]Interval, 0, minAlloc),
		values: make(map[Interval]map[K]T, minAlloc),
	}
}
