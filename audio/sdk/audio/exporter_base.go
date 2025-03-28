package audio

import (
	"context"
	"errors"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"log/slog"

	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/errs"

	"github.com/zalgonoise/x/audio/fft"
)

const (
	errDomain = errs.Domain("x/audio/sdk/audio")

	ErrNil = errs.Kind("nil")

	ErrEmitter    = errs.Entity("emitter")
	ErrCollectors = errs.Entity("collectors")
)

var (
	ErrNilEmitter    = errs.WithDomain(errDomain, ErrNil, ErrEmitter)
	ErrNilCollectors = errs.WithDomain(errDomain, ErrNil, ErrCollectors)
)

type ExporterMetrics interface{}

type exporter struct {
	logger  *slog.Logger
	metrics ExporterMetrics
	tracer  trace.Tracer

	peaks    Collector[float64]
	spectrum Collector[[]fft.FrequencyPower]

	emitter Emitter

	cancel context.CancelFunc
}

// Export implements the Exporter interface.
//
// It consumes the audio data chunks from the Processor, as the signal is streamed from a Process call.
//
// It will use the configured Collector types to extract meaningful data from the signal (peaks and spectrum data),
// as a series of steps in a Collector pipeline, usually composed of an Extractor, a Registry and optionally
// a Compactor.
//
// The returned error is a wrapped error of both peaks and spectrum Collector Collect method call, if raised.
func (e exporter) Export(ctx context.Context, h *wav.Header, data []float64) error {
	e.logger.DebugContext(ctx, "exporting audio chunks from processor",
		slog.Int("num_samples", len(data)),
	)

	return errors.Join(
		e.peaks.Collect(ctx, h, data),
		e.spectrum.Collect(ctx, h, data),
	)
}

// ForceFlush implements the Exporter and StreamCloser interfaces.
//
// It will call on the peaks and spectrum Collector ForceFlush method if their Registry has it.
//
// The returned error is a wrapped error of both peaks and spectrum Collector ForceFlush method call, if raised.
func (e exporter) ForceFlush() error {
	e.logger.DebugContext(context.TODO(), "flushing exporter registry buffers")

	return errors.Join(
		e.peaks.ForceFlush(),
		e.spectrum.ForceFlush(),
	)
}

// Shutdown implements the Exporter and StreamCloser interfaces.
//
// It will stop the running goroutine which listens to the Registry's incoming values. Then, it will call on the
// peaks and spectrum Collector Shutdown method if their Extractor has it, as well as their Registry's Shutdown method.
// Lastly, its Emitter is gracefully shut down via its Shutdown method as well.
//
// The returned error is a wrapped error of both peaks and spectrum Collector ForceFlush method call, if raised.
func (e exporter) Shutdown(ctx context.Context) error {
	e.logger.InfoContext(context.TODO(), "shutting down exporter")

	e.cancel()

	return errors.Join(
		e.peaks.Shutdown(ctx),
		e.spectrum.Shutdown(ctx),
		e.emitter.Shutdown(ctx),
	)
}

func (e exporter) export(ctx context.Context) {
	peaksValues := e.peaks.Load()
	spectrumValues := e.spectrum.Load()

	for {
		select {
		case <-ctx.Done():
			return
		case v, ok := <-peaksValues:
			if !ok {
				e.logger.DebugContext(ctx, "peaks values channel closed")

				return
			}

			e.logger.DebugContext(ctx, "received peaks values")

			e.emitter.EmitPeaks(v)
		case v, ok := <-spectrumValues:
			if !ok {
				e.logger.DebugContext(ctx, "spectrum values channel closed")

				return
			}

			e.logger.DebugContext(ctx, "received spectrum values")

			e.emitter.EmitSpectrum(v)
		}
	}
}

// NewExporter creates an audio Exporter based on the input Emitter (that publishes the values somewhere), and
// the input peaks and spectrum Collector, which will extract, process and / or buffer values from the incoming
// audio signal.
//
// The input Emitter and both Collector types need to be configured on their own, if they require any custom
// configuration, preparation or connection between each other. For example, the caller may want to create a
// Map-Reduce-like pipeline -- which is possible by combining the right Extractor, and a batch Registry with a
// configured Compactor. While this covers the Collector; the Emitter should exclusively be concerned with publishing
// these values to the appropriate destination and / or backend.
//
// A nil Emitter results in a no-op Exporter with an ErrNilEmitter error. Also, if both peaks and spectrum Collector
// types are nil, a no-op Exporter and ErrNilCollectors error are returned. Otherwise, the Exporter is created,
// using a no-op Collector where the provided one is nil.
//
// The created Exporter launches a goroutine to listen on both Collector types' value channel, that is controlled via
// a context.Context. Its done signal is sent on the Exporter's Shutdown method call.
func NewExporter(
	emitter Emitter, peaks Collector[float64], spectrum Collector[[]fft.FrequencyPower],
	logger *slog.Logger, metrics ProcessorMetrics, tracer trace.Tracer,
) (Exporter, error) {
	switch {
	case emitter == nil:
		return NoOpExporter(), ErrNilEmitter
	case peaks == nil && spectrum == nil:
		return NoOpExporter(), ErrNilCollectors
	case peaks == nil:
		peaks = NoOpCollector[float64]()
	case spectrum == nil:
		spectrum = NoOpCollector[[]fft.FrequencyPower]()
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

	ctx, cancel := context.WithCancel(context.Background())

	e := exporter{
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,

		peaks:    peaks,
		spectrum: spectrum,
		emitter:  emitter,
		cancel:   cancel,
	}

	go e.export(ctx)

	return e, nil
}
