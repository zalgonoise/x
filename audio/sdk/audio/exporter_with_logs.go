package audio

import (
	"context"
	"log/slog"
)

type loggedExporter struct {
	e      Exporter
	logger *slog.Logger
}

// Export implements the Exporter interface.
//
// This call will simply execute the inner Exporter's Export method.
func (e loggedExporter) Export(h Header, data []float64) error { return e.e.Export(h, data) }

// ForceFlush implements the Exporter and StreamCloser interfaces.
//
// This call will execute the inner Exporter's ForceFlush method, registering a log event before it does, and
// a Warn log event if it raises an error
func (e loggedExporter) ForceFlush() error {
	ctx := context.Background()
	e.logger.InfoContext(ctx, "flushing")

	if err := e.e.ForceFlush(); err != nil {
		e.logger.WarnContext(ctx, "error when flushing", slog.String("error", err.Error()))

		return err
	}

	return nil
}

// Shutdown implements the Exporter, Closer and StreamCloser interfaces.
//
// This call will execute the inner Exporter's Shutdown method, registering a log event before it does, and
// a Warn log event if it raises an error
func (e loggedExporter) Shutdown(ctx context.Context) error {
	e.logger.InfoContext(ctx, "shutting down Exporter")

	if err := e.e.Shutdown(ctx); err != nil {
		e.logger.WarnContext(ctx, "failed to gracefully shut down", slog.String("error", err.Error()))

		return err
	}

	return nil
}

// ExporterWithLogs returns an Exporter wrapped or decorated with a logger
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
