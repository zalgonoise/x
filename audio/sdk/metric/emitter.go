package metric

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"sync"

	"github.com/zalgonoise/x/audio/sdk"
)

// EmitFunc adapts a function to the Emitter interface, with no-op ForceFlush and Shutdown.
type EmitFunc func(ctx context.Context, batch []Measurement) error

// Emit implements Emitter.
func (f EmitFunc) Emit(ctx context.Context, batch []Measurement) error { return f(ctx, batch) }

// ForceFlush implements Emitter.
func (EmitFunc) ForceFlush(context.Context) error { return nil }

// Shutdown implements Emitter.
func (EmitFunc) Shutdown(context.Context) error { return nil }

// SlogEmitter logs every Measurement as one record named after its instrument, with the channel, frame window and
// sample rate as attributes, plus "value" for a scalar, or "frequencies" and "values" (two arrays, in ascending
// frequency) for a spectrum. Values that are not finite are logged as the strings "NaN", "+Inf" and "-Inf", which JSON
// can hold.
type SlogEmitter struct {
	logger *slog.Logger
	level  slog.Level

	mu     sync.Mutex
	closed bool
}

// NewSlogEmitter returns an Emitter logging to logger at level.
func NewSlogEmitter(logger *slog.Logger, level slog.Level) (*SlogEmitter, error) {
	if logger == nil {
		return nil, fmt.Errorf("%w: nil logger", sdk.ErrInvalidOption)
	}

	return &SlogEmitter{logger: logger, level: level}, nil
}

// Emit implements Emitter.
func (e *SlogEmitter) Emit(ctx context.Context, batch []Measurement) error {
	e.mu.Lock()
	closed := e.closed
	e.mu.Unlock()

	if closed {
		return sdk.ErrShutdown
	}

	for _, m := range batch {
		attrs := []slog.Attr{
			slog.Int("channel", m.Channel),
			slog.Int64("start_frame", m.Start),
			slog.Int64("end_frame", m.End),
			slog.Int("sample_rate", m.SampleRate),
		}

		if m.Kind == KindSpectrum {
			attrs = append(attrs, spectrumAttrs(m.Spectrum)...)
		} else {
			attrs = append(attrs, slog.Attr{Key: "value", Value: floatValue(m.Value)})
		}

		e.logger.LogAttrs(ctx, e.level, m.Name, attrs...)
	}

	return nil
}

// spectrumAttrs returns a spectrum as two arrays of the same length: the bins' centre frequencies and their values.
func spectrumAttrs(bins []Bin) []slog.Attr {
	frequencies, values := make([]float64, len(bins)), make([]any, len(bins))

	for i, b := range bins {
		frequencies[i], values[i] = b.Frequency, floatValue(b.Value).Any()
	}

	return []slog.Attr{slog.Any("frequencies", frequencies), slog.Any("values", values)}
}

// floatValue returns v as a number, or as a string when it is not finite: JSON has no NaN or infinities.
func floatValue(v float64) slog.Value {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return slog.StringValue(strconv.FormatFloat(v, 'g', -1, 64))
	}

	return slog.Float64Value(v)
}

// ForceFlush implements Emitter; records are written as they are logged.
func (e *SlogEmitter) ForceFlush(context.Context) error { return nil }

// Shutdown implements Emitter: later Emit calls return sdk.ErrShutdown. It is idempotent.
func (e *SlogEmitter) Shutdown(context.Context) error {
	e.mu.Lock()
	e.closed = true
	e.mu.Unlock()

	return nil
}
