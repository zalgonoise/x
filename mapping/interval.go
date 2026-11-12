// Package mapping provides time-range interval mapping and conflict-resolution
// primitives for merging, splitting, and organizing values across overlapping
// timeframes.
//
// Intervals are represented by [Interval] and paired with arbitrary data via
// [DataInterval]. Multiple intervals can be organized in batch using [Organize],
// [OrganizeSeq], or [Timeframe], or merged eagerly on insertion using
// [TimeframeStream].
//
// Conflict-resolution strategies are configured with functional options such as
// [WithReplace], [WithFlatten] (or [WithMergeFunc]), [WithKeep], [WithIntersect],
// and [WithAppend], alongside [WithCompareFunc] and [WithOffset].
package mapping

import (
	"iter"
	"time"
)

const minAlloc = 64

// Interval represents a period of time bounded by From and To.
type Interval struct {
	From time.Time
	To   time.Time
}

// IsZero reports whether both From and To are the zero [time.Time] instant.
func (i Interval) IsZero() bool {
	return i.From.IsZero() && i.To.IsZero()
}

// IsValid reports whether the interval has at least one non-zero bound and
// From is not after To.
func (i Interval) IsValid() bool {
	return !i.IsZero() && !i.To.Before(i.From)
}

// Duration returns the elapsed time between From and To. If From is after To,
// Duration returns 0.
func (i Interval) Duration() time.Duration {
	if i.To.Before(i.From) {
		return 0
	}

	return i.To.Sub(i.From)
}

// Compare compares i and other chronologically by From, breaking ties by To.
// It returns -1 if i is before other, +1 if i is after other, and 0 if they have
// equal bounds.
func (i Interval) Compare(other Interval) int {
	if cmp := i.From.Compare(other.From); cmp != 0 {
		return cmp
	}

	return i.To.Compare(other.To)
}

// Overlaps reports whether i and other share a non-empty intersection
// (i.To > other.From and i.From < other.To).
func (i Interval) Overlaps(other Interval) bool {
	return i.To.After(other.From) && i.From.Before(other.To)
}

// Contains reports whether t falls within the inclusive bounds [From, To].
func (i Interval) Contains(t time.Time) bool {
	return !t.Before(i.From) && !t.After(i.To)
}

// DataInterval associates a value of type T with an [Interval].
type DataInterval[T any] struct {
	Data     T
	Interval Interval
}

// AsSeq returns an [iter.Seq2] iterator yielding the [Interval] and value of
// each element in data in slice order.
func AsSeq[T any](data []DataInterval[T]) iter.Seq2[Interval, T] {
	return func(yield func(Interval, T) bool) {
		for i := range data {
			if !yield(data[i].Interval, data[i].Data) {
				return
			}
		}
	}
}

// Collect gathers all (Interval, T) pairs from seq into a newly allocated slice
// of [DataInterval]. If seq is nil, Collect returns nil.
func Collect[T any](seq iter.Seq2[Interval, T]) []DataInterval[T] {
	if seq == nil {
		return nil
	}

	out := make([]DataInterval[T], 0, minAlloc)

	for interval, value := range seq {
		out = append(out, DataInterval[T]{
			Interval: interval,
			Data:     value,
		})
	}

	return out
}
