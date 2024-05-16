package mapping

import "github.com/zalgonoise/cfg"

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
func Organize[M TimeframeType[T, K], T any, K any](seq SeqKV[Interval, T], reducer ReducerFunc[T]) *K {
	flattened := reducer(seq)

	var tf M = new(K)

	tf = tf.init()

	flattened(func(interval Interval, t T) bool {
		_ = tf.Add(interval, t)

		return true
	})

	return tf
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
