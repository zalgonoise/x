package mapping

import (
	"slices"
	"time"
)

// Interval is a period of time with a From and To time.Time values.
type Interval struct {
	From time.Time
	To   time.Time
}

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

// OrganizeSeq consumes a sequence Seq of Interval and data, returning an appropriate timeframe type K with organized
// / flattened values.
func OrganizeSeq[M TimeframeType[T, K], T any, K any](seq SeqKV[Interval, T], reducer ReducerFunc[T]) *K {
	flattened := reducer(seq)

	var tf M = new(K)

	tf = tf.init()

	flattened(func(interval Interval, t T) bool {
		_ = tf.Add(interval, t)

		return true
	})

	return tf
}

func Organize[T comparable](
	data []DataInterval[T], mergeFunc func(a, b T) T, offset time.Duration,
) []DataInterval[T] {
	cache := make([]DataInterval[T], 0, minAlloc)

	for i := range data {
		cache = resolveConflicts(data[i].Interval, data[i].Data, cache, split, mergeFunc, offset)
	}

	slices.SortFunc(cache, func(a, b DataInterval[T]) int {
		return a.Interval.From.Compare(b.Interval.From)
	})

	return cache
}

func OrganizeFunc[T any](
	data []DataInterval[T], cmpFunc func(a, b T) bool, mergeFunc func(a, b T) T, offset time.Duration,
) []DataInterval[T] {
	cache := make([]DataInterval[T], 0, minAlloc)

	for i := range data {
		cache = resolveAnyConflicts(data[i].Interval, data[i].Data, cache, split, cmpFunc, mergeFunc, offset)
	}

	slices.SortFunc(cache, func(a, b DataInterval[T]) int {
		return a.Interval.From.Compare(b.Interval.From)
	})

	return cache
}
