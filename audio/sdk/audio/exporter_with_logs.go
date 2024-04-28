package audio

import (
	"context"
	"log/slog"

	"github.com/zalgonoise/x/audio/encoding/wav"
)

type loggedExporter struct {
	e      Exporter
	logger *slog.Logger
}

// Export implements the Exporter interface.
//
// This implementation calls the underlying Exporter's Export method without taking any other action.
//
// It consumes the audio data chunks from the Processor, as the signal is streamed from a Process call.
//
// It will use the configured Collector types to extract meaningful data from the signal (peaks and spectrum data),
// as a series of steps in a Collector pipeline, usually composed of an Extractor, a Registry and optionally
// a Compactor.
//
// The returned error is a wrapped error of both peaks and spectrum Collector Collect method call, if raised.
func (e loggedExporter) Export(ctx context.Context, h *wav.Header, data []float64) error {
	return e.e.Export(ctx, h, data)
}

// ForceFlush implements the Exporter and StreamCloser interfaces.
//
// This implementation calls the underlying Exporter's ForceFlush method, registering log entries before the
// call and if it raises an error with a Warn-level event.
//
// It will call on the peaks and spectrum Collector ForceFlush method if their Registry has it.
//
// The returned error is a wrapped error of both peaks and spectrum Collector ForceFlush method call, if raised.
func (e loggedExporter) ForceFlush() error {
	ctx := context.Background()
	e.logger.InfoContext(ctx, "flushing")

	if err := e.e.ForceFlush(); err != nil {
		e.logger.WarnContext(ctx, "error when flushing", slog.String("error", err.Error()))

		return err
	}

	return nil
}

// Shutdown implements the Exporter and StreamCloser interfaces.
//
// This implementation calls the underlying Exporter's Shutdown method, registering log entries before the
// call and if it raises an error with a Warn-level event.
//
// It will stop the running goroutine which listens to the Registry's incoming values. Then, it will call on the
// peaks and spectrum Collector Shutdown method if their Extractor has it, as well as their Registry's Shutdown method.
// Lastly, its Emitter is gracefully shut down via its Shutdown method as well.
//
// The returned error is a wrapped error of both peaks and spectrum Collector ForceFlush method call, if raised.
func (e loggedExporter) Shutdown(ctx context.Context) error {
	e.logger.InfoContext(ctx, "shutting down Exporter")

	if err := e.e.Shutdown(ctx); err != nil {
		e.logger.WarnContext(ctx, "failed to gracefully shut down", slog.String("error", err.Error()))

		return err
	}

	return nil
}

// ExporterWithLogs decorates the input Exporter with a slog.Logger using the input slog.Handler.
//
// If the Exporter is nil, a no-op Exporter is returned. If the input slog.Handler is nil, a default
// text handler is created as a safe default. If the input Exporter is already a logged Exporter; then
// its logger's handler is replaced with this handler (input or default one).
//
// This Exporter will not add any new functionality besides decorating the Exporter with log events.
func ExporterWithLogs(e Exporter, handler slog.Handler) Exporter {
	if e == nil {
		return NoOpExporter()
	}

	if handler == nil {
		handler = newDefaultHandler()
	}

	if withLogs, ok := (e).(loggedExporter); ok {
		withLogs.logger = slog.New(handler)

		return withLogs
	}

	return loggedExporter{
		e:      e,
		logger: slog.New(handler),
	}
}
