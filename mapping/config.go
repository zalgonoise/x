package mapping

import (
	"reflect"
	"time"

	"github.com/zalgonoise/cfg"
)

type strategy uint8

const (
	strategyReplace strategy = iota
	strategyFlatten
	strategyKeep
	strategyIntersect
	strategyAppend
)

// MergeFunc defines how two values are combined when an incoming interval (next)
// overlaps with an existing interval (cur).
type MergeFunc[T any] func(cur, next T) T

// CompareFunc reports whether two values of type T should be considered equal
// when coalescing adjacent intervals.
type CompareFunc[T any] func(a, b T) bool

// MergeReplace is a pre-built [MergeFunc] that overwrites the existing value
// with the incoming value (last-write-wins).
func MergeReplace[T any](_, next T) T {
	return next
}

// MergeKeep is a pre-built [MergeFunc] that preserves the existing value and
// ignores the incoming value on overlapping segments (first-write-wins).
func MergeKeep[T any](cur, _ T) T {
	return cur
}

// Config holds the configuration parameters for organizing and merging
// intervals in a [Timeframe], [TimeframeStream], [Organize], or [OrganizeSeq].
type Config[T any] struct {
	strategy  strategy
	mergeFunc MergeFunc[T]
	cmpFunc   CompareFunc[T]
	offset    time.Duration
}

func defaultConfig[T any]() Config[T] {
	return configure(Config[T]{})
}

func configure[T any](base Config[T], opts ...cfg.Option[Config[T]]) Config[T] {
	c := base
	if len(opts) > 0 {
		c = cfg.Set(c, opts...)
	}

	if c.mergeFunc == nil {
		c.mergeFunc = MergeReplace[T]
	}

	if c.cmpFunc == nil {
		if rt := reflect.TypeFor[T](); rt != nil && rt.Comparable() {
			c.cmpFunc = func(a, b T) bool {
				return any(a) == any(b)
			}
		}
	}

	return c
}

// WithReplace configures the conflict-resolution strategy to overwrite
// overlapping sub-intervals with the incoming (next) value.
//
// WithReplace is the default strategy when no strategy option is provided.
func WithReplace[T any]() cfg.Option[Config[T]] {
	return cfg.Register(func(config Config[T]) Config[T] {
		config.strategy = strategyReplace
		config.mergeFunc = MergeReplace[T]

		return config
	})
}

// WithKeep configures the conflict-resolution strategy to preserve the existing
// (cur) value on overlapping sub-intervals, only inserting non-overlapping
// portions of incoming intervals.
func WithKeep[T any]() cfg.Option[Config[T]] {
	return cfg.Register(func(config Config[T]) Config[T] {
		config.strategy = strategyKeep
		config.mergeFunc = MergeKeep[T]

		return config
	})
}

// WithFlatten configures the conflict-resolution strategy to split overlapping
// intervals and combine their values on intersecting segments using mergeFunc.
//
// If mergeFunc is nil, [MergeReplace] is used.
func WithFlatten[T any](mergeFunc MergeFunc[T]) cfg.Option[Config[T]] {
	return cfg.Register(func(config Config[T]) Config[T] {
		config.strategy = strategyFlatten
		if mergeFunc == nil {
			config.mergeFunc = MergeReplace[T]
		} else {
			config.mergeFunc = mergeFunc
		}

		return config
	})
}

// WithMergeFunc is an alias for [WithFlatten] that configures conflict
// resolution using the provided [MergeFunc] (such as a custom function,
// [MergeReplace], or [MergeKeep]).
func WithMergeFunc[T any](mergeFunc MergeFunc[T]) cfg.Option[Config[T]] {
	return WithFlatten(mergeFunc)
}

// WithIntersect configures the conflict-resolution strategy to retain only the
// intersecting sub-intervals where an incoming interval overlaps existing
// intervals, combining their values using mergeFunc.
//
// If mergeFunc is nil, [MergeReplace] is used.
func WithIntersect[T any](mergeFunc MergeFunc[T]) cfg.Option[Config[T]] {
	return cfg.Register(func(config Config[T]) Config[T] {
		config.strategy = strategyIntersect
		if mergeFunc == nil {
			config.mergeFunc = MergeReplace[T]
		} else {
			config.mergeFunc = mergeFunc
		}

		return config
	})
}

// WithAppend configures the organizer to append intervals and sort them
// chronologically by (From, To) without splitting or merging overlaps.
func WithAppend[T any]() cfg.Option[Config[T]] {
	return cfg.Register(func(config Config[T]) Config[T] {
		config.strategy = strategyAppend

		return config
	})
}

// WithCompareFunc configures a custom [CompareFunc] to determine whether two
// adjacent intervals carry equal values and should be coalesced into a single
// continuous interval.
//
// For comparable types T, standard Go equality (==) is used automatically when
// WithCompareFunc is not specified.
func WithCompareFunc[T any](cmpFunc CompareFunc[T]) cfg.Option[Config[T]] {
	return cfg.Register(func(config Config[T]) Config[T] {
		if cmpFunc != nil {
			config.cmpFunc = cmpFunc
		}

		return config
	})
}

// WithOffset configures a discrete time offset applied at split boundaries
// between adjacent intervals (for example, time.Second when working with
// inclusive second-granularity timestamps).
//
// By default, offset is 0, treating intervals as contiguous half-open ranges
// [From, To) where adjacent intervals share an exact boundary timestamp.
func WithOffset[T any](offset time.Duration) cfg.Option[Config[T]] {
	return cfg.Register(func(config Config[T]) Config[T] {
		if offset >= 0 {
			config.offset = offset
		}

		return config
	})
}
