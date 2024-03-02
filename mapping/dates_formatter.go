package mapping

import (
	"time"

	"github.com/zalgonoise/cfg"
)

type TimeSeq[K comparable, T any] interface {
	Add(i Interval, values map[K]T) bool
	All() SeqKV[Interval, map[K]T]
}

type TimeFormatter[K comparable, T any] struct {
	fnFrom func(time.Time) (time.Time, bool)
	fnTo   func(time.Time) (time.Time, bool)

	tf TimeSeq[K, T]
}

func NewTimeFormatter[K comparable, T any](tf TimeSeq[K, T], opts ...cfg.Option[Format]) TimeSeq[K, T] {
	format := cfg.New(opts...)

	if format.fnFrom == nil && format.fnTo == nil {
		return tf
	}

	return TimeFormatter[K, T]{
		fnFrom: format.fnFrom,
		fnTo:   format.fnTo,
		tf:     tf,
	}
}

func (t TimeFormatter[K, T]) Add(i Interval, values map[K]T) bool {
	var ok bool

	if t.fnFrom != nil {
		i.From, ok = t.fnFrom(i.From)
		if !ok {
			return false
		}
	}

	if t.fnTo != nil {
		i.To, ok = t.fnTo(i.To)
		if !ok {
			return false
		}
	}

	return t.tf.Add(i, values)
}

func (t TimeFormatter[K, T]) All() SeqKV[Interval, map[K]T] {
	return t.tf.All()
}
