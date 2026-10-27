package writer

import (
	"context"
	"sync"
	"time"

	"github.com/zalgonoise/cfg"
	"github.com/zalgonoise/x/gacha"
)

// request is one Write call queued for a Batch flush.
type request[R any] struct {
	data []R
	err  chan error
}

// Batch is a gacha.Writer that coalesces concurrent Write calls into shared Sink.Write calls (group commit), which
// cuts the number of round-trips and transactions under load. It is safe for concurrent use.
//
// Each call's items are marshaled by the caller, then flushed by a background goroutine together with the items of
// other calls, in a single Sink.Write. Every caller blocks until its batch is persisted and gets that batch's error.
// The items of a single call are never split across flushes, so each call is atomic.
//
// A Batch must be closed with Close to flush pending writes and stop its goroutine.
type Batch[T, R any] struct {
	sink   gacha.Sink[R]
	m      gacha.Marshaler[T, R]
	config Config

	reqs     chan *request[R]
	closing  chan struct{}
	done     chan struct{}
	close    sync.Once
	requests sync.Pool

	// Owned by the flush goroutine.
	batch []*request[R]
	data  []R
}

// NewBatch creates a Batch writer over sink, marshaling items with m, and starts its flush goroutine. Both must be
// non-nil. It accepts the WithMaxBatchSize, WithLinger, WithFlushTimeout, WithRetry and WithNotify options.
func NewBatch[T, R any](sink gacha.Sink[R], m gacha.Marshaler[T, R], opts ...cfg.Option[Config]) *Batch[T, R] {
	b := &Batch[T, R]{
		sink:    sink,
		m:       m,
		config:  cfg.Set(defaultBatchConfig(), opts...),
		reqs:    make(chan *request[R]),
		closing: make(chan struct{}),
		done:    make(chan struct{}),
	}

	go b.run()

	return b
}

// Write marshals items and queues them for the next flush, then waits until that flush persisted them. Nothing is
// queued if an item fails to marshal.
//
// If ctx is done before the items are queued, Write returns the context's error and nothing is written. If ctx is
// done while waiting for the flush, Write returns the context's error, but the items may still be persisted.
// Write returns ErrClosed after Close was called.
func (b *Batch[T, R]) Write(ctx context.Context, items ...T) error {
	if len(items) == 0 {
		return nil
	}

	select {
	case <-b.closing:
		return ErrClosed
	default:
	}

	req := b.getRequest(len(items))

	data, err := marshal(b.m, req.data, items)
	req.data = data

	if err != nil {
		b.putRequest(req)

		return err
	}

	select {
	case b.reqs <- req:
	case <-b.closing:
		b.putRequest(req)

		return ErrClosed
	case <-ctx.Done():
		b.putRequest(req)

		return ctx.Err()
	}

	select {
	case err = <-req.err:
		b.putRequest(req)

		return err
	case <-ctx.Done():
		// The request is queued and will still be flushed: it cannot be reused.
		return ctx.Err()
	}
}

// Close stops accepting writes, flushes the pending ones and waits for the flush goroutine to exit, or for ctx to be
// done. It is safe to call Close more than once.
func (b *Batch[T, R]) Close(ctx context.Context) error {
	b.close.Do(func() {
		close(b.closing)
	})

	select {
	case <-b.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (b *Batch[T, R]) run() {
	defer close(b.done)

	var carry *request[R]

	for {
		first := carry
		if first == nil {
			first = b.next()
			if first == nil {
				return
			}
		}

		b.batch = append(b.batch[:0], first)
		carry = b.collect(len(first.data))

		b.flush()
	}
}

// next waits for the first request of a batch. It returns nil once the writer is closed and no request is pending.
func (b *Batch[T, R]) next() *request[R] {
	select {
	case req := <-b.reqs:
		return req
	case <-b.closing:
		select {
		case req := <-b.reqs:
			return req
		default:
			return nil
		}
	}
}

// collect adds pending requests to the batch until it is full, the linger time is up, or (without linger) no request
// is pending. It returns the request that would have overflowed the batch, if any, to start the next one with.
func (b *Batch[T, R]) collect(size int) *request[R] {
	var linger <-chan time.Time

	if b.config.linger > 0 {
		timer := time.NewTimer(b.config.linger)
		defer timer.Stop()

		linger = timer.C
	}

	for b.config.maxBatchSize == 0 || size < b.config.maxBatchSize {
		var req *request[R]

		if linger == nil {
			select {
			case req = <-b.reqs:
			default:
				return nil
			}
		} else {
			select {
			case req = <-b.reqs:
			case <-linger:
				return nil
			case <-b.closing:
				// Stop lingering, but keep collecting what is already pending.
				linger = nil

				continue
			}
		}

		if b.config.maxBatchSize > 0 && size+len(req.data) > b.config.maxBatchSize {
			return req
		}

		b.batch = append(b.batch, req)
		size += len(req.data)
	}

	return nil
}

func (b *Batch[T, R]) flush() {
	data := b.data[:0]

	for _, req := range b.batch {
		data = append(data, req.data...)
	}

	ctx := context.Background()

	if b.config.flushTimeout > 0 {
		var cancel context.CancelFunc

		ctx, cancel = context.WithTimeout(ctx, b.config.flushTimeout)
		defer cancel()
	}

	err := write(ctx, b.sink, data, &b.config)

	clear(data)
	b.data = data[:0]

	for _, req := range b.batch {
		req.err <- err
	}

	clear(b.batch)
	b.batch = b.batch[:0]

	if err == nil {
		notify(b.config.notify)
	}
}

func (b *Batch[T, R]) getRequest(n int) *request[R] {
	if req, ok := b.requests.Get().(*request[R]); ok && cap(req.data) >= n {
		return req
	}

	return &request[R]{
		data: make([]R, 0, n),
		err:  make(chan error, 1),
	}
}

func (b *Batch[T, R]) putRequest(req *request[R]) {
	clear(req.data)
	req.data = req.data[:0]

	b.requests.Put(req)
}
