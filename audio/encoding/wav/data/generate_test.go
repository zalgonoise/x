package data_test

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/encoding/wav/data"
	"github.com/zalgonoise/x/audio/osc"
)

// TestGenerateIsNormalized guards against oscillator output scaled to the bit depth (±32767) being stored in the
// normalized float buffer, which made every generated file clip.
func TestGenerateIsNormalized(t *testing.T) {
	for _, depth := range []uint16{8, 16, 24, 32} {
		chunk := data.NewPCMChunk(depth, nil)
		chunk.Generate(osc.SineWave, 1000, 48000, 10*time.Millisecond)

		peak := 0.0
		for _, v := range chunk.Float() {
			peak = math.Max(peak, math.Abs(v))
		}

		require.InDelta(t, 1.0, peak, 1e-3, "%d-bit: a full-scale sine peaks at 1.0", depth)

		values := chunk.Value()
		maxV := 1<<(depth-1) - 1
		require.Equal(t, 0, values[0], "%d-bit: the sine starts at zero", depth)

		for i, v := range values {
			require.LessOrEqual(t, v, maxV, "%d-bit sample %d", depth, i)
			require.GreaterOrEqual(t, v, -maxV-1, "%d-bit sample %d", depth, i)
		}
	}

	ring := data.NewPCMRing(16, nil, 4096, func([]float64) error { return nil })
	ring.Generate(osc.SineWave, 1000, 48000, 10*time.Millisecond)

	for _, v := range ring.Float() {
		require.LessOrEqual(t, math.Abs(v), 1.0)
	}
}
