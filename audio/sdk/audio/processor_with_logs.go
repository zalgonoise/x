package audio

import (
	"context"
	"io"
	"log/slog"
)

type loggedProcessor struct {
	p      Processor
	logger *slog.Logger
}

// Process decorates a Processor interface implementation.
//
// This implementation calls the underlying Processor's Process method, registering log entries before the
// call and when it is terminated.
//
// It reads the byte stream from the input io.Reader using the configured StreamExporter, which
// is both converting the bytes data into floating point audio; processing it, and finally exporting
// it to its configured outputs.
//
// Process should be called in a goroutine, as a blocking call that is supposed to be issued asynchronously.
func (p loggedProcessor) Process(ctx context.Context, reader io.Reader) {
	p.logger.InfoContext(ctx, "processing signal from streamer")

	p.p.Process(ctx, reader)

	p.logger.InfoContext(ctx, "terminating Process call")
}

// Err decorates a Processor interface implementation.
//
// This implementation will simply return a call to the underlying Processor's Err method.
//
// It returns a receiving channel for errors, that allows the caller of a Process method to listen for any raised
// errors.
func (p loggedProcessor) Err() <-chan error {
	return p.p.Err()
}

// ForceFlush decorates a Processor interface implementation.
//
// This implementation calls the underlying Processor's ForceFlush method, registering log entries before the
// call and if it raises an error with a Warn-level event.
//
// It allows direct access to the configured Exporter's (or, StreamExporter's) ForceFlush method, for example,
// when its exporter contains a batched Registry that buffers items in chunks.
func (p loggedProcessor) ForceFlush() error {
	ctx := context.Background()
	p.logger.InfoContext(ctx, "flushing")

	if err := p.p.ForceFlush(); err != nil {
		p.logger.WarnContext(ctx, "error when flushing", slog.String("error", err.Error()))

		return err
	}

	return nil
}

// Shutdown implements the Processor and StreamCloser interfaces.
//
// This implementation calls the underlying Processor's Shutdown method, registering log entries before the
// call and if it raises an error with a Warn-level event.
//
// It allows a graceful shutdown of the Processor. It starts by stopping its runtime, and then gracefully shutting down
// the configured Exporter (or, StreamExporter).
func (p loggedProcessor) Shutdown(ctx context.Context) error {
	p.logger.InfoContext(ctx, "shutting down Processor")

	if err := p.p.Shutdown(ctx); err != nil {
		p.logger.WarnContext(ctx, "failed to gracefully shut down", slog.String("error", err.Error()))

		return err
	}

	return nil
}

// ProcessorWithLogs decorates the input Processor with a slog.Logger using the input slog.Handler.
//
// If the Processor is nil, a no-op Processor is returned. If the input slog.Handler is nil, a default
// text handler is created as a safe default. If the input Processor is already a logged Processor; then
// its logger's handler is replaced with this handler (input or default one).
//
// This Processor will not add any new functionality besides decorating the Processor with log events.
func ProcessorWithLogs(p Processor, handler slog.Handler) Processor {
	if p == nil {
		return NoOpProcessor()
	}

	if handler == nil {
		handler = newDefaultHandler()
	}

	if withLogs, ok := (p).(loggedProcessor); ok {
		withLogs.logger = slog.New(handler)

		return withLogs
	}

	return loggedProcessor{
		p:      p,
		logger: slog.New(handler),
	}
}
