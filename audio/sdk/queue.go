package sdk

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	otelmetric "go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"

	"github.com/zalgonoise/x/audio"
)

// Policy decides what a QueuedExporter does with a new Buffer when its queue is full.
type Policy uint8

const (
	// DropOldest discards the oldest queued Buffer to make room, so a slow exporter catches up with the latest audio
	// and the producer never waits. It is the default (CLAUDE.md §7 Q8): right for live streams and analysis.
	DropOldest Policy = iota
	// Block makes Export wait for room (or for its context), so nothing is lost: right for recording and for
	// transcoding files, which decode much faster than real time.
	Block
)

// String returns the policy's name.
func (p Policy) String() string {
	switch p {
	case DropOldest:
		return "drop-oldest"
	case Block:
		return "block"
	default:
		return fmt.Sprintf("policy(%d)", uint8(p))
	}
}

// DefaultQueueSize is how many Buffers a QueuedExporter holds unless WithQueueSize says otherwise.
const DefaultQueueSize = 64

// QueueOption configures a QueuedExporter.
type QueueOption func(*queueConfig) error

type queueConfig struct {
	size    int
	policy  Policy
	onError ErrorHandler
	meter   otelmetric.MeterProvider
}

// WithQueueSize sets how many Buffers the queue holds (at least 1). The default is DefaultQueueSize.
func WithQueueSize(n int) QueueOption {
	return func(c *queueConfig) error {
		if n < 1 {
			return fmt.Errorf("%w: queue size %d", ErrInvalidOption, n)
		}

		c.size = n

		return nil
	}
}

// WithPolicy sets what happens when the queue is full. The default is DropOldest.
func WithPolicy(p Policy) QueueOption {
	return func(c *queueConfig) error {
		if p != DropOldest && p != Block {
			return fmt.Errorf("%w: %s", ErrInvalidOption, p)
		}

		c.policy = p

		return nil
	}
}

// WithQueueErrorHandler sets the handler for errors returned by the wrapped Exporter's Export, which runs
// asynchronously. The default discards them.
func WithQueueErrorHandler(fn ErrorHandler) QueueOption {
	return func(c *queueConfig) error {
		if fn == nil {
			return fmt.Errorf("%w: nil error handler", ErrInvalidOption)
		}

		c.onError = fn

		return nil
	}
}

// WithQueueMeterProvider sets where the queue's audio.queue.dropped counter goes (Buffers dropped by DropOldest or
// abandoned by Shutdown). The default is a no-op provider.
func WithQueueMeterProvider(mp otelmetric.MeterProvider) QueueOption {
	return func(c *queueConfig) error {
		if mp == nil {
			return fmt.Errorf("%w: nil meter provider", ErrInvalidOption)
		}

		c.meter = mp

		return nil
	}
}

// QueuedExporter decouples an Exporter from its producer, like the OpenTelemetry BatchSpanProcessor: Export copies the
// Buffer into a bounded queue and returns, and one goroutine feeds the queue to the wrapped Exporter in order.
type QueuedExporter struct {
	next Exporter
	cfg  queueConfig

	mu      sync.Mutex
	items   []queued
	buffers int  // data items in items, the ones that count against the size
	closing bool // no new items; the worker drains and exits

	wake   chan struct{} // the worker may have something to do
	space  chan struct{} // a data item left the queue
	closed chan struct{} // Shutdown started: blocked producers give up
	done   chan struct{} // the worker exited

	ctx    context.Context // for the wrapped Exporter's Export, cancelled when Shutdown gives up waiting
	cancel context.CancelFunc

	dropped        atomic.Uint64
	droppedCounter otelmetric.Int64Counter

	shutdownOnce sync.Once
	shutdownDone chan struct{}
	shutdownErr  error
}

// queued is a Buffer to export, or a flush request when flushed is not nil.
type queued struct {
	buf     audio.Buffer
	flushed chan error
	flushTo context.Context //nolint:containedctx // the flush caller's context, used by the worker
}

// NewQueuedExporter starts a QueuedExporter in front of next.
func NewQueuedExporter(next Exporter, opts ...QueueOption) (*QueuedExporter, error) {
	if next == nil {
		return nil, ErrNilExporter
	}

	cfg := queueConfig{
		size:    DefaultQueueSize,
		policy:  DropOldest,
		onError: func(error) {},
		meter:   metricnoop.NewMeterProvider(),
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}

		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}

	dropped, err := cfg.meter.Meter(instrumentationName).Int64Counter("audio.queue.dropped",
		otelmetric.WithUnit("{buffer}"), otelmetric.WithDescription("Buffers dropped by a QueuedExporter"))
	if err != nil {
		return nil, fmt.Errorf("sdk: creating a counter: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	q := &QueuedExporter{
		droppedCounter: dropped,
		next:           next,
		cfg:            cfg,
		wake:           make(chan struct{}, 1),
		space:          make(chan struct{}, 1),
		closed:         make(chan struct{}),
		done:           make(chan struct{}),
		ctx:            ctx,
		cancel:         cancel,
		shutdownDone:   make(chan struct{}),
	}

	go q.run()

	return q, nil
}

// Dropped returns how many Buffers the DropOldest policy has discarded, and Shutdown has abandoned.
func (q *QueuedExporter) Dropped() uint64 { return q.dropped.Load() }

// Export queues a copy of b. With DropOldest it never waits; with Block it waits for room until ctx is done. It
// returns ErrShutdown after Shutdown.
func (q *QueuedExporter) Export(ctx context.Context, b audio.Buffer) error {
	b = b.Clone()

	for {
		q.mu.Lock()

		switch {
		case q.closing:
			q.mu.Unlock()

			return ErrShutdown
		case q.buffers < q.cfg.size:
			q.items = append(q.items, queued{buf: b})
			q.buffers++
			q.mu.Unlock()
			signal(q.wake)

			return nil
		case q.cfg.policy == DropOldest:
			q.dropOldest()
			q.items = append(q.items, queued{buf: b})
			q.mu.Unlock()
			signal(q.wake)

			return nil
		}

		q.mu.Unlock()

		select {
		case <-q.space:
		case <-q.closed:
			return ErrShutdown
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// dropOldest removes the oldest data item; the caller holds mu and keeps the count, as it adds one back.
func (q *QueuedExporter) dropOldest() {
	for i, it := range q.items {
		if it.flushed == nil {
			q.items = append(q.items[:i:i], q.items[i+1:]...)
			q.drop()

			return
		}
	}
}

// ForceFlush waits until every Buffer queued before the call was exported, then flushes the wrapped Exporter.
func (q *QueuedExporter) ForceFlush(ctx context.Context) error {
	flushed := make(chan error, 1)

	q.mu.Lock()

	if q.closing {
		q.mu.Unlock()

		return ErrShutdown
	}

	q.items = append(q.items, queued{flushed: flushed, flushTo: ctx})
	q.mu.Unlock()
	signal(q.wake)

	select {
	case err := <-flushed:
		return err
	case <-ctx.Done():
		q.unqueueFlush(flushed) // an abandoned request must not pile up, nor flush later with a dead ctx

		return ctx.Err()
	}
}

// unqueueFlush removes a flush request that is still queued.
func (q *QueuedExporter) unqueueFlush(flushed chan error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, it := range q.items {
		if it.flushed == flushed {
			q.items = append(q.items[:i:i], q.items[i+1:]...)

			return
		}
	}
}

// Shutdown stops accepting Buffers, waits for the queue to drain into the wrapped Exporter, and shuts it down. If ctx
// ends first, the export in progress is cancelled, the rest of the queue is dropped (and counted), and ctx's error is
// returned along with the wrapped Exporter's. It is idempotent and safe for concurrent use; the first caller's ctx
// governs the drain.
func (q *QueuedExporter) Shutdown(ctx context.Context) error {
	q.shutdownOnce.Do(func() {
		q.mu.Lock()
		q.closing = true
		q.mu.Unlock()
		close(q.closed)
		signal(q.wake)

		go func() {
			var drainErr error

			select {
			case <-q.done:
			case <-ctx.Done():
				drainErr = ctx.Err()

				q.cancel()
				<-q.done
			}

			q.cancel()
			q.shutdownErr = errors.Join(drainErr, q.next.Shutdown(ctx))
			close(q.shutdownDone)
		}()
	})

	return awaitResult(ctx, q.shutdownDone, func() error { return q.shutdownErr })
}

// run is the worker: the only goroutine calling the wrapped Exporter's Export and ForceFlush.
func (q *QueuedExporter) run() {
	defer close(q.done)

	for {
		it, ok := q.pop()
		if !ok {
			return
		}

		if it.flushed != nil {
			if err := it.flushTo.Err(); err != nil { // its caller gave up already
				it.flushed <- err

				continue
			}

			it.flushed <- q.next.ForceFlush(it.flushTo)

			continue
		}

		if err := q.next.Export(q.ctx, it.buf); err != nil {
			q.cfg.onError(fmt.Errorf("sdk: queued export: %w", err))
		}
	}
}

// pop waits for the next item. It reports false once the queue is closing and empty, or abandoned by Shutdown.
func (q *QueuedExporter) pop() (queued, bool) {
	for {
		q.mu.Lock()

		if q.ctx.Err() != nil { // Shutdown gave up waiting: drop what is left
			for _, it := range q.items {
				if it.flushed != nil {
					it.flushed <- ErrShutdown
				} else {
					q.drop()
				}
			}

			q.items, q.buffers = nil, 0
			q.mu.Unlock()

			return queued{}, false
		}

		if len(q.items) > 0 {
			it := q.items[0]
			q.items[0] = queued{}
			q.items = q.items[1:]

			if it.flushed == nil {
				q.buffers--
			}

			q.mu.Unlock()
			signal(q.space)

			return it, true
		}

		closing := q.closing
		q.mu.Unlock()

		if closing {
			return queued{}, false
		}

		select {
		case <-q.wake:
		case <-q.ctx.Done():
		}
	}
}

// drop counts a dropped Buffer.
func (q *QueuedExporter) drop() {
	q.dropped.Add(1)
	q.droppedCounter.Add(context.Background(), 1)
}

// signal wakes a waiter without blocking; a pending signal is enough.
func signal(c chan struct{}) {
	select {
	case c <- struct{}{}:
	default:
	}
}
