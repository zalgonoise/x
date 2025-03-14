package audio

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/trace/noop"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/zalgonoise/x/errs"
	"go.opentelemetry.io/otel/trace"
)

const (
	ErrHalt = errs.Kind("process stopped")

	ErrSignal = errs.Entity("with OS signal")
)

// ErrHaltSignal is a sentinel error for when an OS signal is emitted, to halt or stop the application's runtime.
var ErrHaltSignal = errs.WithDomain(errDomain, ErrHalt, ErrSignal)

type ProcessorMetrics interface{}

type processor struct {
	logger  *slog.Logger
	metrics ProcessorMetrics
	tracer  trace.Tracer

	streamExporter StreamExporter

	errCh  chan error
	cancel context.CancelFunc
}

// Process implements the Processor interface.
//
// It reads the byte stream from the input io.Reader using the configured StreamExporter, which
// is both converting the bytes data into floating point audio; processing it, and finally exporting
// it to its configured outputs.
//
// Process should be called in a goroutine, as a blocking call that is supposed to be issued asynchronously.
func (p *processor) Process(ctx context.Context, reader io.Reader) {
	p.logger.InfoContext(ctx, "processing signal from streamer")

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	defer func() {
		close(signalCh)

		p.logger.InfoContext(ctx, "terminating Process call")
	}()

	ctx, p.cancel = context.WithCancel(ctx)

	go p.streamExporter.Stream(ctx, reader, p.errCh)

	for {
		select {
		case <-ctx.Done():
			if e := ctx.Err(); e != nil {
				p.logger.DebugContext(ctx, "closing on context done", slog.String("error", e.Error()))
			}

			p.cancel()

			return
		case sig := <-signalCh:
			p.logger.DebugContext(ctx, "received halt signal")

			p.errCh <- fmt.Errorf("%w: %s", ErrHaltSignal, sig.String())
			p.cancel()

			return
		}
	}
}

// Err implements the Processor interface.
//
// It returns a receiving channel for errors, that allows the caller of a Process method to listen for any raised
// errors.
func (p *processor) Err() <-chan error {
	return p.errCh
}

// ForceFlush implements the Processor and StreamCloser interfaces.
//
// It allows direct access to the configured Exporter's (or, StreamExporter's) ForceFlush method, for example,
// when its exporter contains a batched Registry that buffers items in chunks.
func (p *processor) ForceFlush() error {
	p.logger.DebugContext(context.TODO(), "flushing processor")

	return p.streamExporter.ForceFlush()
}

// Shutdown implements the Processor and StreamCloser interfaces.
//
// It allows a graceful shutdown of the Processor. It starts by stopping its runtime, and then gracefully shutting down
// the configured Exporter (or, StreamExporter).
func (p *processor) Shutdown(ctx context.Context) error {
	p.logger.InfoContext(ctx, "shutting down Processor")

	//TODO: why does it panic on closing a closed channel? Double-shutdown where?
	// defer close(p.errCh)
	p.cancel()

	return p.streamExporter.Shutdown(ctx)
}

// NewProcessor creates a Processor from a StreamExporter.
//
// The input StreamExporter should be already set-up and ready to be used, since the Processor will simply
// call on its available methods. It is the responsibility of the caller to set it up accordingly, especially when
// both Streamer and Exporter types are connected in any way.
func NewProcessor(
	streamExporter StreamExporter,
	logger *slog.Logger, metrics ProcessorMetrics, tracer trace.Tracer,
) Processor {
	if streamExporter == nil {
		return NoOpProcessor()
	}

	if logger == nil {
		logger = slog.New(noOpLogHandler{})
	}

	if metrics == nil {
		metrics = NoOpProcessorMetrics{}
	}

	if tracer == nil {
		tracer = noop.NewTracerProvider().Tracer("no-op")
	}

	return &processor{
		logger:         logger,
		metrics:        metrics,
		tracer:         tracer,
		errCh:          make(chan error),
		streamExporter: streamExporter,
	}
}
