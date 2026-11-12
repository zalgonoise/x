package osc_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/osc"
)

// TestSineStartsAtPhaseZero guards against the off-by-one in the table lookup that duplicated sample 0 and delayed
// every waveform by one sample.
func TestSineStartsAtPhaseZero(t *testing.T) {
	const (
		sampleRate = 44100
		freq       = 1000
		depth      = 16
	)

	buf := make([]float64, 64)
	osc.Sine(buf, freq, depth, sampleRate)

	scale := float64(int(1)<<(depth-1) - 1)

	for n, got := range buf {
		want := scale * math.Sin(2*math.Pi*freq*float64(n)/sampleRate)
		require.InDelta(t, want, got, 1e-6*scale, "sample %d", n)
	}
}
