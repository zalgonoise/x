package mapping

import (
	"iter"
	"slices"

	"github.com/zalgonoise/cfg"
)

// Timeframe is a batch container for (Interval, T) pairs.
//
// Intervals added via [Timeframe.Add], [Timeframe.AddSlice], or [Timeframe.Append]
// are buffered in insertion order and resolved into a new organized [Timeframe]
// when [Timeframe.Organize] is called.
type Timeframe[T any] struct {
	buffer  []DataInterval[T]
	config  Config[T]
	readIdx int64
}

// NewTimeframe creates a new [Timeframe] of type T configured with opts.
//
// Options provided to NewTimeframe serve as defaults when [Timeframe.Organize]
// is called without overriding options.
func NewTimeframe[T any](opts ...cfg.Option[Config[T]]) *Timeframe[T] {
	return &Timeframe[T]{
		buffer: make([]DataInterval[T], 0, minAlloc),
		config: configure(defaultConfig[T](), opts...),
	}
}

func (t *Timeframe[T]) init() {
	if t.buffer == nil {
		t.buffer = make([]DataInterval[T], 0, minAlloc)
	}

	if t.config.mergeFunc == nil {
		t.config = defaultConfig[T]()
	}
}

// Add appends the [Interval] i and value to the [Timeframe] buffer.
// It returns false if t is nil or if i is not a valid interval.
func (t *Timeframe[T]) Add(i Interval, value T) bool {
	if t == nil || !i.IsValid() {
		return false
	}

	t.init()
	t.buffer = append(t.buffer, DataInterval[T]{
		Interval: i,
		Data:     value,
	})

	return true
}

// AddSlice appends all items into the [Timeframe] buffer.
// It returns false if t is nil or if any item in items has an invalid interval
// (valid items are still added).
func (t *Timeframe[T]) AddSlice(items ...DataInterval[T]) bool {
	if t == nil {
		return false
	}

	t.init()

	ok := true
	for i := range items {
		if !t.Add(items[i].Interval, items[i].Data) {
			ok = false
		}
	}

	return ok
}

// Append iterates through seq and adds each (Interval, T) pair to the [Timeframe].
// It returns false if t or seq is nil, or if any interval in seq is invalid.
func (t *Timeframe[T]) Append(seq iter.Seq2[Interval, T]) bool {
	if t == nil || seq == nil {
		return false
	}

	t.init()

	ok := true
	for interval, value := range seq {
		if !t.Add(interval, value) {
			ok = false
		}
	}

	return ok
}

// All returns an [iter.Seq2] iterator over the (Interval, T) pairs currently
// stored in the [Timeframe].
func (t *Timeframe[T]) All() iter.Seq2[Interval, T] {
	if t == nil {
		return func(func(Interval, T) bool) {}
	}

	return AsSeq(t.buffer)
}

// Slice returns a shallow copy of the [DataInterval] slice stored in the [Timeframe].
func (t *Timeframe[T]) Slice() []DataInterval[T] {
	if t == nil || len(t.buffer) == 0 {
		return []DataInterval[T]{}
	}

	return slices.Clone(t.buffer)
}

// Len returns the number of intervals currently stored in the [Timeframe].
func (t *Timeframe[T]) Len() int {
	if t == nil {
		return 0
	}

	return len(t.buffer)
}

// Organize resolves conflicts across all buffered intervals using the strategy
// and options configured on t (overridden by any opts passed to Organize),
// returning a new organized [Timeframe].
func (t *Timeframe[T]) Organize(opts ...cfg.Option[Config[T]]) *Timeframe[T] {
	if t == nil {
		return NewTimeframe[T](opts...)
	}

	t.init()
	config := configure(t.config, opts...)

	return &Timeframe[T]{
		buffer: organizeSlice(t.buffer, config),
		config: config,
	}
}

// Merge combines the intervals of t and other, returning a new organized
// [Timeframe] using t's configuration (overridden by any opts).
func (t *Timeframe[T]) Merge(other *Timeframe[T], opts ...cfg.Option[Config[T]]) *Timeframe[T] {
	if t == nil && other == nil {
		return NewTimeframe[T](opts...)
	}

	if t == nil {
		return other.Organize(opts...)
	}

	t.init()
	config := configure(t.config, opts...)

	if other == nil || len(other.buffer) == 0 {
		return &Timeframe[T]{
			buffer: organizeSlice(t.buffer, config),
			config: config,
		}
	}

	combined := make([]DataInterval[T], 0, len(t.buffer)+len(other.buffer))
	combined = append(combined, t.buffer...)
	combined = append(combined, other.buffer...)

	return &Timeframe[T]{
		buffer: organizeSlice(combined, config),
		config: config,
	}
}

// TimeframeStream is a streaming container for (Interval, T) pairs that
// resolves conflicts and maintains chronological order eagerly on every
// [TimeframeStream.Add], [TimeframeStream.AddSlice], or [TimeframeStream.Append] call.
type TimeframeStream[T any] struct {
	buffer  []DataInterval[T]
	scratch []DataInterval[T]
	config  Config[T]
	readIdx int64
	seeded  bool
}

// NewTimeframeStream creates a new [TimeframeStream] of type T configured with opts.
//
// By default, if no strategy option is provided, [WithReplace] is used.
func NewTimeframeStream[T any](opts ...cfg.Option[Config[T]]) *TimeframeStream[T] {
	return &TimeframeStream[T]{
		buffer: make([]DataInterval[T], 0, minAlloc),
		config: configure(Config[T]{}, opts...),
	}
}

func (t *TimeframeStream[T]) init() {
	if t.buffer == nil {
		t.buffer = make([]DataInterval[T], 0, minAlloc)
	}

	if t.config.mergeFunc == nil {
		t.config = defaultConfig[T]()
	}
}

// Add inserts the [Interval] i and value into the [TimeframeStream], immediately
// resolving any overlaps with previously added intervals according to the
// stream's configured strategy.
//
// It returns false if t is nil or if i is not a valid interval.
func (t *TimeframeStream[T]) Add(i Interval, value T) bool {
	if t == nil {
		return false
	}

	t.init()

	if !isValidInterval(i, t.config.offset) {
		return false
	}

	t.buffer, t.scratch = resolveStep(t.buffer, t.scratch, i, value, t.config, t.seeded)
	t.seeded = true

	return true
}

// AddSlice inserts each item in items into the [TimeframeStream], resolving
// conflicts eagerly on each insertion.
//
// It returns false if t is nil or if any item in items has an invalid interval.
func (t *TimeframeStream[T]) AddSlice(items ...DataInterval[T]) bool {
	if t == nil {
		return false
	}

	t.init()

	ok := true
	for i := range items {
		if !t.Add(items[i].Interval, items[i].Data) {
			ok = false
		}
	}

	return ok
}

// Append iterates through seq and adds each (Interval, T) pair to the
// [TimeframeStream], resolving conflicts eagerly on each insertion.
//
// It returns false if t or seq is nil, or if any interval in seq is invalid.
func (t *TimeframeStream[T]) Append(seq iter.Seq2[Interval, T]) bool {
	if t == nil || seq == nil {
		return false
	}

	t.init()

	ok := true
	for interval, value := range seq {
		if !t.Add(interval, value) {
			ok = false
		}
	}

	return ok
}

// All returns an [iter.Seq2] iterator over the organized (Interval, T) pairs
// in the [TimeframeStream], ordered chronologically by From.
func (t *TimeframeStream[T]) All() iter.Seq2[Interval, T] {
	if t == nil {
		return func(func(Interval, T) bool) {}
	}

	return AsSeq(t.buffer)
}

// Slice returns a shallow copy of the organized [DataInterval] slice in the
// [TimeframeStream].
func (t *TimeframeStream[T]) Slice() []DataInterval[T] {
	if t == nil || len(t.buffer) == 0 {
		return []DataInterval[T]{}
	}

	return slices.Clone(t.buffer)
}

// Len returns the number of organized intervals currently stored in the
// [TimeframeStream].
func (t *TimeframeStream[T]) Len() int {
	if t == nil {
		return 0
	}

	return len(t.buffer)
}

// Organize returns the [TimeframeStream] itself, as its intervals are already
// organized on insertion.
func (t *TimeframeStream[T]) Organize() *TimeframeStream[T] {
	return t
}
