package audio

import (
	"context"

	"github.com/zalgonoise/x/audio/fft"
)

type noOpEmitter struct{}

// EmitPeaks implements the Emitter interface.
//
// This is a no-op call, and has no effect.
func (noOpEmitter) EmitPeaks(float64) {}

// EmitSpectrum implements the Emitter interface.
//
// This is a no-op call, and has no effect.
func (noOpEmitter) EmitSpectrum([]fft.FrequencyPower) {}

// Shutdown implements the Emitter and Closer interfaces.
//
// This is a no-op call, and the returned error is always nil.
func (noOpEmitter) Shutdown(context.Context) error { return nil }

// NoOpEmitter returns a no-op Emitter.
func NoOpEmitter() Emitter {
	return noOpEmitter{}
}
