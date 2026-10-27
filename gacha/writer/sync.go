package writer

import (
	"context"
	"sync"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/gacha"
)

// Sync is a gacha.Writer that marshals the items of each Write call and writes them to its sink right away, in one
// Sink.Write call (or one per chunk, with WithMaxBatchSize). It is safe for concurrent use.
type Sync[T, R any] struct {
	sink   gacha.Sink[R]
	m      gacha.Marshaler[T, R]
	config Config
	raws   sync.Pool
}

// NewSync creates a Sync writer over sink, marshaling items with m. Both must be non-nil. It accepts the
// WithMaxBatchSize, WithRetry and WithNotify options.
func NewSync[T, R any](sink gacha.Sink[R], m gacha.Marshaler[T, R], opts ...cfg.Option[Config]) *Sync[T, R] {
	return &Sync[T, R]{
		sink:   sink,
		m:      m,
		config: cfg.Set(defaultSyncConfig(), opts...),
	}
}

// Write marshals items and writes them to the sink, returning once they are persisted. Nothing is written if an item
// fails to marshal. Without WithMaxBatchSize, the items are written atomically. With it, they are written in chunks:
// a failure leaves the earlier chunks persisted.
func (w *Sync[T, R]) Write(ctx context.Context, items ...T) error {
	if len(items) == 0 {
		return nil
	}

	raws := w.getRaws(len(items))
	defer w.putRaws(raws)

	data, err := marshal(w.m, (*raws)[:0], items)
	*raws = data

	if err != nil {
		return err
	}

	size := w.config.maxBatchSize
	if size <= 0 {
		size = len(data)
	}

	var written bool

	for chunk := range chunks(data, size) {
		if err = write(ctx, w.sink, chunk, &w.config); err != nil {
			break
		}

		written = true
	}

	if written {
		notify(w.config.notify)
	}

	return err
}

func (w *Sync[T, R]) getRaws(n int) *[]R {
	if raws, ok := w.raws.Get().(*[]R); ok && cap(*raws) >= n {
		return raws
	}

	raws := make([]R, 0, n)

	return &raws
}

func (w *Sync[T, R]) putRaws(raws *[]R) {
	// Drop the references to the raw values, so pooled slices do not keep them alive.
	clear(*raws)
	*raws = (*raws)[:0]

	w.raws.Put(raws)
}

// chunks yields consecutive sub-slices of data with at most size elements.
func chunks[R any](data []R, size int) func(yield func([]R) bool) {
	return func(yield func([]R) bool) {
		for len(data) > 0 {
			n := min(size, len(data))
			if !yield(data[:n:n]) {
				return
			}

			data = data[n:]
		}
	}
}
