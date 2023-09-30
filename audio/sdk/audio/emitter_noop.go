package audio

import (
	"context"

	"github.com/zalgonoise/x/audio/fft"
)

type noOpEmitter struct{}

func (noOpEmitter) EmitPeaks(float64)                 {}
func (noOpEmitter) EmitSpectrum([]fft.FrequencyPower) {}
func (noOpEmitter) Shutdown(context.Context) error    { return nil }

func NoOpEmitter() Emitter {
	return noOpEmitter{}
}
