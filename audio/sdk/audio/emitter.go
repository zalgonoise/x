package audio

import (
	"github.com/zalgonoise/x/audio/fft"
)

type Emitter interface {
	EmitPeaks(float64)
	EmitSpectrum([]fft.FrequencyPower)

	Closer
}
