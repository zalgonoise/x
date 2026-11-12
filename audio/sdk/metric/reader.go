package metric

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/zalgonoise/x/audio/sdk"
)

// Errors returned by Readers.
var (
	// ErrRegistered is returned when a second Meter registers with a Reader: each Reader serves one Meter.
	ErrRegistered = errors.New("metric: reader already serves a meter")
	// ErrNoMeter is returned by ManualReader.Collect before a Meter registered with it.
	ErrNoMeter = errors.New("metric: no meter registered")
	// ErrEmitterPanic wraps a panic recovered from an Emitter, which would otherwise crash the process.
	ErrEmitterPanic = errors.New("metric: emitter panicked")
)

// Reader collects a Meter's measurements. It is implemented by PeriodicReader and ManualReader.
type Reader interface {
	ForceFlush(ctx context.Context) error
	Shutdown(ctx context.Context) error

	register(collect func() []Measurement) error
}

// producer holds the collect function a Meter registered.
type producer struct {
	mu      sync.Mutex
	collect func() []Measurement
}

func (p *producer) register(collect func() []Measurement) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.collect != nil {
		return ErrRegistered
	}

	p.collect = collect

	return nil
}

func (p *producer) produce() ([]Measurement, bool) {
	p.mu.Lock()
	collect := p.collect
	p.mu.Unlock()

	if collect == nil {
		return nil, false
	}

	return collect(), true
}

// ManualReader collects on demand, for pull-based backends such as a Prometheus scrape.
type ManualReader struct {
	producer

	mu     sync.Mutex
	closed bool
}

// NewManualReader returns a ManualReader.
func NewManualReader() *ManualReader { return &ManualReader{} }

// Collect returns the measurements recorded since the previous Collect. It returns ErrNoMeter before a Meter uses the
// reader, and sdk.ErrShutdown after Shutdown.
func (r *ManualReader) Collect(context.Context) ([]Measurement, error) {
	r.mu.Lock()
	closed := r.closed
	r.mu.Unlock()

	if closed {
		return nil, sdk.ErrShutdown
	}

	batch, ok := r.produce()
	if !ok {
		return nil, ErrNoMeter
	}

	return batch, nil
}

// ForceFlush does nothing: a ManualReader only collects when asked.
func (r *ManualReader) ForceFlush(context.Context) error { return nil }

// Shutdown makes later Collect calls return sdk.ErrShutdown. It is idempotent.
func (r *ManualReader) Shutdown(context.Context) error {
	r.mu.Lock()
	r.closed = true
	r.mu.Unlock()

	return nil
}

// Emitter sends batches of measurements somewhere: logs, stdout, an OTLP endpoint. Emit must not keep batch.
type Emitter interface {
	Emit(ctx context.Context, batch []Measurement) error
	ForceFlush(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

// DefaultInterval is how often a PeriodicReader collects unless WithInterval says otherwise.
const DefaultInterval = time.Second

// PeriodicOption configures a PeriodicReader.
type PeriodicOption func(*PeriodicReader) error

// WithInterval sets how often the reader collects and emits. The default is DefaultInterval.
func WithInterval(d time.Duration) PeriodicOption {
	return func(r *PeriodicReader) error {
		if d <= 0 {
			return fmt.Errorf("%w: interval %s", sdk.ErrInvalidOption, d)
		}

		r.interval = d

		return nil
	}
}

// WithEmitTimeout bounds each periodic Emit call. The default is the interval.
func WithEmitTimeout(d time.Duration) PeriodicOption {
	return func(r *PeriodicReader) error {
		if d <= 0 {
			return fmt.Errorf("%w: emit timeout %s", sdk.ErrInvalidOption, d)
		}

		r.timeout = d

		return nil
	}
}

// WithErrorHandler sets the handler for errors from periodic Emit calls. The default discards them.
func WithErrorHandler(fn sdk.ErrorHandler) PeriodicOption {
	return func(r *PeriodicReader) error {
		if fn == nil {
			return fmt.Errorf("%w: nil error handler", sdk.ErrInvalidOption)
		}

		r.onError = fn

		return nil
	}
}

// PeriodicReader collects every interval and hands non-empty batches to an Emitter, from one goroutine that owns all
// calls to Emit. It starts when a Meter registers with it.
type PeriodicReader struct {
	producer

	emitter  Emitter
	interval time.Duration
	timeout  time.Duration
	onError  sdk.ErrorHandler

	flush chan flushRequest
	stop  chan context.Context // a final emit, then exit
	quit  chan struct{}        // exit at once (Shutdown's ctx ended first)
	done  chan error

	mu           sync.Mutex
	started      bool
	closed       bool
	shutdownOnce sync.Once
	shutdownDone chan struct{}
	shutdownErr  error
}

// flushRequest asks the reader's goroutine to emit now, bounded by the caller's ctx.
type flushRequest struct {
	ctx   context.Context //nolint:containedctx // the flush caller's context, used by the reader's goroutine
	reply chan error
}

// NewPeriodicReader returns a PeriodicReader emitting to emitter.
func NewPeriodicReader(emitter Emitter, opts ...PeriodicOption) (*PeriodicReader, error) {
	if emitter == nil {
		return nil, fmt.Errorf("%w: nil emitter", sdk.ErrInvalidOption)
	}

	r := &PeriodicReader{
		emitter:      emitter,
		interval:     DefaultInterval,
		onError:      func(error) {},
		flush:        make(chan flushRequest),
		stop:         make(chan context.Context),
		quit:         make(chan struct{}),
		done:         make(chan error, 1),
		shutdownDone: make(chan struct{}),
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}

		if err := opt(r); err != nil {
			return nil, err
		}
	}

	if r.timeout == 0 {
		r.timeout = r.interval
	}

	return r, nil
}

func (r *PeriodicReader) register(collect func() []Measurement) error {
	if err := r.producer.register(collect); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return sdk.ErrShutdown
	}

	r.started = true

	go r.run()

	return nil
}

func (r *PeriodicReader) run() {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), r.timeout)

			if err := r.emit(ctx); err != nil {
				r.onError(fmt.Errorf("metric: periodic emit: %w", err))
			}

			cancel()
		case req := <-r.flush:
			req.reply <- r.emit(req.ctx)
		case ctx := <-r.stop:
			r.done <- r.emit(ctx)

			return
		case <-r.quit:
			return
		}
	}
}

// emit collects and emits one batch; empty batches are not emitted. A panicking Emitter is reported as an error.
func (r *PeriodicReader) emit(ctx context.Context) (err error) {
	batch, ok := r.produce()
	if !ok || len(batch) == 0 {
		return nil
	}

	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("%w: %v", ErrEmitterPanic, p)
		}
	}()

	return r.emitter.Emit(ctx, batch)
}

// ForceFlush collects and emits now, then flushes the Emitter.
func (r *PeriodicReader) ForceFlush(ctx context.Context) error {
	r.mu.Lock()
	closed, started := r.closed, r.started
	r.mu.Unlock()

	switch {
	case closed:
		return sdk.ErrShutdown
	case !started:
		return r.emitter.ForceFlush(ctx)
	}

	reply := make(chan error, 1)

	select {
	case r.flush <- flushRequest{ctx: ctx, reply: reply}:
	case <-r.quit:
		return sdk.ErrShutdown
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case err := <-reply:
		return errors.Join(err, r.emitter.ForceFlush(ctx))
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Shutdown collects and emits one last batch, stops the reader and shuts the Emitter down. It is idempotent and safe
// for concurrent use: every call waits for the same shutdown up to its own ctx; the first caller's ctx bounds the final
// emit.
func (r *PeriodicReader) Shutdown(ctx context.Context) error {
	r.shutdownOnce.Do(func() {
		r.mu.Lock()
		r.closed = true
		started := r.started
		r.mu.Unlock()

		go func() {
			var emitErr error

			if started {
				select {
				case r.stop <- ctx:
					emitErr = <-r.done
				case <-ctx.Done():
					emitErr = ctx.Err()
				}
			}

			close(r.quit) // whatever happened above, the goroutine exits

			r.shutdownErr = errors.Join(emitErr, r.emitter.Shutdown(ctx))
			close(r.shutdownDone)
		}()
	})

	return awaitResult(ctx, r.shutdownDone, func() error { return r.shutdownErr })
}
