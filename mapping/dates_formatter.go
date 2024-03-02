package mapping

import (
	"time"

	"github.com/zalgonoise/cfg"
)

type TimeField[K comparable, T any] interface {
	Add(i Interval, values map[K]T) bool
	Append(seq SeqKV[Interval, map[K]T]) (err error)
	All() SeqKV[Interval, map[K]T]
}

type TimeFormatter[K comparable, T any] struct {
	fnFrom func(time.Time) (time.Time, bool)
	fnTo   func(time.Time) (time.Time, bool)

	tf TimeField[K, T]
}

func NewTimeFormatter[K comparable, T any](tf TimeField[K, T], opts ...cfg.Option[Format]) TimeField[K, T] {
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
	var (
		from = i.From
		to   = i.To
		ok   bool
	)

	if t.fnFrom != nil {
		from, ok = t.fnFrom(i.From)
		if !ok {
			return false
		}
	}

	if t.fnTo != nil {
		to, ok = t.fnTo(i.To)
		if !ok {
			return false
		}
	}

	return t.tf.Add(Interval{
		From: from,
		To:   to,
	}, values)
}

func (t TimeFormatter[K, T]) Append(seq SeqKV[Interval, map[K]T]) (err error) {
	if !seq(t.Add) {
		return errAppendFailed
	}

	return nil
}

func (t TimeFormatter[K, T]) All() SeqKV[Interval, map[K]T] {
	return t.tf.All()
}
