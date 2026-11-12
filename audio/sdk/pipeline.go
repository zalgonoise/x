package sdk

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/zalgonoise/x/audio"
)

const (
	// maxBufferSamples bounds the samples in one Buffer (8 MiB of float64), whatever the duration and format ask for.
	maxBufferSamples = 1 << 20
	// maxEmptyReads is how many consecutive empty reads fill tolerates, as bufio does.
	maxEmptyReads = 100
)

// Pipeline reads a Source, decodes it and fans every Buffer out to its Exporters, in order, on the Run goroutine.
//
// Exporters see each Buffer synchronously, so a slow one slows the whole pipeline down; wrap it in a QueuedExporter to
// decouple it. Errors from Export go to the ErrorHandler and do not stop the pipeline. Errors from the Source, the
// decoder or a Transform end Run.
//
// A Pipeline can Run again after a Run returned (to reconnect to a live source, for example), but not concurrently,
// and not after Shutdown.
type Pipeline struct {
	src       Source
	exporters []Exporter
	cfg       config

	tracer       trace.Tracer
	buffers      otelmetric.Int64Counter
	frames       otelmetric.Int64Counter
	exportErrors otelmetric.Int64Counter

	mu       sync.Mutex
	running  bool
	cancel   context.CancelCauseFunc
	done     chan struct{} // closed when the current Run returns
	shutdown bool

	shutdownOnce sync.Once
	shutdownDone chan struct{}
	shutdownErr  error
}

// NewPipeline returns a Pipeline reading src and exporting to exporters.
func NewPipeline(src Source, exporters []Exporter, opts ...Option) (*Pipeline, error) {
	if src == nil {
		return nil, ErrNilSource
	}

	for i, e := range exporters {
		if e == nil {
			return nil, fmt.Errorf("%w: exporter %d", ErrNilExporter, i)
		}
	}

	cfg, err := newConfig(opts)
	if err != nil {
		return nil, err
	}

	p := &Pipeline{
		src:          src,
		exporters:    append([]Exporter(nil), exporters...),
		cfg:          cfg,
		tracer:       cfg.tracer.Tracer(instrumentationName),
		shutdownDone: make(chan struct{}),
	}

	meter := cfg.meter.Meter(instrumentationName)

	if p.buffers, err = meter.Int64Counter("audio.pipeline.buffers", otelmetric.WithUnit("{buffer}"),
		otelmetric.WithDescription("Buffers handed to the exporters")); err != nil {
		return nil, fmt.Errorf("sdk: creating a counter: %w", err)
	}

	if p.frames, err = meter.Int64Counter("audio.pipeline.frames", otelmetric.WithUnit("{frame}"),
		otelmetric.WithDescription("Frames handed to the exporters")); err != nil {
		return nil, fmt.Errorf("sdk: creating a counter: %w", err)
	}

	if p.exportErrors, err = meter.Int64Counter("audio.pipeline.export.errors", otelmetric.WithUnit("{error}"),
		otelmetric.WithDescription("Export calls that returned an error")); err != nil {
		return nil, fmt.Errorf("sdk: creating a counter: %w", err)
	}

	return p, nil
}

// Run opens the Source and exports its audio until the stream ends, and then returns nil. Each Buffer carries the
// configured duration (see WithBufferDuration), except the last one, so a live stream's latency is one buffer. It returns ErrShutdown when
// Shutdown stopped it, the context's error when ctx was cancelled, and any error from the Source, the decoder or a
// Transform. Cancelling ctx or calling Shutdown closes the source stream, which unblocks a pending read.
//
// Run does not shut the Exporters down: call Shutdown (or ForceFlush) when done with the Pipeline.
func (p *Pipeline) Run(ctx context.Context) (err error) {
	ctx, finish, err := p.start(ctx)
	if err != nil {
		return err
	}

	defer finish()

	ctx, span := p.tracer.Start(ctx, "audio.pipeline.run")

	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		span.End()
	}()

	stream, err := p.src.Open(ctx)
	if err != nil {
		return p.stopped(ctx, fmt.Errorf("sdk: opening the source: %w", err))
	}

	var closeOnce sync.Once

	closeStream := func() { closeOnce.Do(func() { _ = stream.Close() }) }

	defer closeStream()

	// a source blocked in Read only returns once its stream is closed
	stopClose := context.AfterFunc(ctx, closeStream)
	defer stopClose()

	sr, err := p.cfg.decode(stream)
	if err != nil {
		return p.stopped(ctx, fmt.Errorf("sdk: decoding the source: %w", err))
	}

	f := sr.Format()
	span.SetAttributes(
		attribute.String("audio.encoding", f.Encoding.String()),
		attribute.Int("audio.sample_rate", f.SampleRate),
		attribute.Int("audio.channels", f.Channels),
		attribute.Int("audio.bit_depth", f.BitDepth),
	)

	return p.stopped(ctx, p.pump(ctx, sr))
}

// start registers a Run, or reports why it cannot start.
func (p *Pipeline) start(parent context.Context) (context.Context, func(), error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	switch {
	case p.shutdown:
		return nil, nil, ErrShutdown
	case p.running:
		return nil, nil, ErrRunning
	}

	ctx, cancel := context.WithCancelCause(parent)
	done := make(chan struct{})

	p.running, p.cancel, p.done = true, cancel, done

	return ctx, func() {
		cancel(nil)

		p.mu.Lock()
		p.running, p.cancel = false, nil
		p.mu.Unlock()

		close(done)
	}, nil
}

// stopped replaces err with the reason Run was stopped, when it was: an error caused by closing the stream is noise.
func (p *Pipeline) stopped(ctx context.Context, err error) error {
	if err != nil && ctx.Err() != nil {
		return context.Cause(ctx)
	}

	return err
}

func (p *Pipeline) pump(ctx context.Context, sr audio.SampleReader) error {
	// the format comes from an untrusted header: it must be valid, and a frame must fit in a bounded buffer
	f := sr.Format()

	if err := f.Validate(); err != nil {
		return fmt.Errorf("sdk: decoder returned %w", err)
	}

	if int64(f.Channels) > maxBufferSamples {
		return fmt.Errorf("sdk: decoder returned %w: %d channels", audio.ErrUnsupportedFormat, f.Channels)
	}

	frames := min(max(f.Frames(p.cfg.bufferDur), 1), maxBufferSamples/int64(f.Channels))
	samples := make([]float64, frames*int64(f.Channels))

	var frame int64

	for {
		n, err := fill(sr, samples)

		if ctx.Err() != nil { // stopped while reading: the samples may come from a stream being torn down
			return context.Cause(ctx)
		}

		if n > 0 {
			b := audio.Buffer{Format: f, Samples: samples[:n], Frame: frame}
			frame += int64(n / f.Channels)

			if err := p.export(ctx, b); err != nil {
				return err
			}
		}

		switch {
		case errors.Is(err, io.EOF):
			return nil
		case err != nil:
			return fmt.Errorf("sdk: reading samples: %w", err)
		case ctx.Err() != nil:
			return context.Cause(ctx)
		}
	}
}

// fill reads whole frames into dst until it is full, or the reader returns an error (with the samples read before it).
// Every Buffer then carries the configured duration, except the last one. A reader returning nothing and no error 100
// times in a row fails with io.ErrNoProgress.
func fill(sr audio.SampleReader, dst []float64) (int, error) {
	n, empty := 0, 0

	for n < len(dst) {
		m, err := sr.ReadSamples(dst[n:])
		n += m

		switch {
		case err != nil:
			return n, err
		case m > 0:
			empty = 0
		default:
			empty++

			if empty >= maxEmptyReads {
				return n, io.ErrNoProgress
			}
		}
	}

	return n, nil
}

func (p *Pipeline) export(ctx context.Context, b audio.Buffer) error {
	for _, t := range p.cfg.transforms {
		var err error

		if b, err = t.Transform(ctx, b); err != nil {
			return fmt.Errorf("sdk: transform: %w", err)
		}
	}

	p.buffers.Add(ctx, 1)
	p.frames.Add(ctx, int64(b.Frames()))

	for _, e := range p.exporters {
		if err := e.Export(ctx, b); err != nil {
			p.exportErrors.Add(ctx, 1)
			p.cfg.onError(fmt.Errorf("sdk: export: %w", err))
		}
	}

	return nil
}

// ForceFlush flushes every Exporter, and returns their errors joined.
func (p *Pipeline) ForceFlush(ctx context.Context) error {
	p.mu.Lock()
	shutdown := p.shutdown
	p.mu.Unlock()

	if shutdown {
		return ErrShutdown
	}

	errs := make([]error, 0, len(p.exporters))
	for _, e := range p.exporters {
		errs = append(errs, e.ForceFlush(ctx))
	}

	return errors.Join(errs...)
}

// Shutdown stops a running Run (which returns ErrShutdown), waits for it to return, and shuts every Exporter down, in
// order, returning their errors joined. It is idempotent and safe for concurrent use: every call waits for the same
// shutdown, up to its own ctx, and a call after it completed returns its result.
//
// The shutdown runs in the background with the first caller's ctx: when that ctx ends before Run returns (a Source
// stuck in Open, for example), Shutdown returns the ctx's error, and the Exporters are still shut down once Run
// returns.
func (p *Pipeline) Shutdown(ctx context.Context) error {
	p.shutdownOnce.Do(func() {
		p.mu.Lock()
		p.shutdown = true
		cancel, done := p.cancel, p.done
		p.mu.Unlock()

		if cancel != nil {
			cancel(ErrShutdown)
		}

		go func() {
			if done != nil {
				<-done
			}

			errs := make([]error, 0, len(p.exporters))
			for _, e := range p.exporters {
				errs = append(errs, e.Shutdown(ctx))
			}

			p.shutdownErr = errors.Join(errs...)
			close(p.shutdownDone)
		}()
	})

	return awaitResult(ctx, p.shutdownDone, func() error { return p.shutdownErr })
}

// awaitResult waits for done and returns result(), or ctx's error if ctx ends first. A result already available wins
// over a ctx that is already done, so repeated calls are deterministic.
func awaitResult(ctx context.Context, done <-chan struct{}, result func() error) error {
	select {
	case <-done:
		return result()
	default:
	}

	select {
	case <-done:
		return result()
	case <-ctx.Done():
		return ctx.Err()
	}
}
