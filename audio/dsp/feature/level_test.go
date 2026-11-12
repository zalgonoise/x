package feature_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/dsp/feature"
	"github.com/zalgonoise/x/audio/internal/testutil"
)

func TestLevels(t *testing.T) {
	const rate, n = 48000, 48000

	sine := testutil.Sine(n, rate, 1000, 0.5, 0) // 1 kHz, amplitude 0.5, whole cycles
	square := make([]float64, n)

	for i := range square {
		square[i] = 0.25
		if (i/24)%2 == 1 { // 1 kHz square wave
			square[i] = -0.25
		}
	}

	require.InDelta(t, 0.5, feature.Peak(sine), 1e-9)
	require.InDelta(t, 0.5/math.Sqrt2, feature.RMS(sine), 1e-9)
	require.InDelta(t, math.Sqrt2, feature.Crest(sine), 1e-6)
	require.InDelta(t, 0, feature.DCOffset(sine), 1e-12)

	require.InDelta(t, 0.25, feature.RMS(square), 1e-12)
	require.InDelta(t, 1, feature.Crest(square), 1e-12)
	require.InDelta(t, 2000, feature.ZeroCrossingRate(square, rate), 1, "two crossings per cycle")
	require.InDelta(t, 2000, feature.ZeroCrossingRate(sine, rate), 2)

	require.InDelta(t, 0.3, feature.DCOffset(testutil.DC(10, 0.3)), 1e-12)
	require.Equal(t, 2, feature.Clipped([]float64{1, -1, 0.99, 0.5}, 1))
	require.Equal(t, 3, feature.Clipped([]float64{1, -1, 0.99, 0.5}, 0.99))

	require.InDelta(t, 0, feature.DBFS(1), 0)
	require.InDelta(t, -6.0206, feature.DBFS(0.5), 1e-4)
	require.True(t, math.IsInf(feature.DBFS(0), -1))
}

func TestEmptyAndEdgeCases(t *testing.T) {
	for _, x := range [][]float64{nil, {}} {
		require.Zero(t, feature.Peak(x))
		require.Zero(t, feature.RMS(x))
		require.Zero(t, feature.Crest(x))
		require.Zero(t, feature.ZeroCrossings(x))
		require.Zero(t, feature.ZeroCrossingRate(x, 48000))
		require.Zero(t, feature.DCOffset(x))
		require.Zero(t, feature.Clipped(x, 1))
	}

	require.Zero(t, feature.Crest(make([]float64, 8)), "silence")
	require.Zero(t, feature.ZeroCrossingRate([]float64{1}, 48000))
	require.Zero(t, feature.ZeroCrossingRate([]float64{1, -1}, 0))
	require.Equal(t, 1, feature.ZeroCrossings([]float64{-1, 0}), "zero counts as positive")
	require.Zero(t, feature.ZeroCrossings([]float64{0, 1}))

	x := []float64{0.5, -0.5}
	feature.Peak(x)
	feature.RMS(x)
	require.Equal(t, []float64{0.5, -0.5}, x, "inputs are never modified")
}

// Review round 3: levels of float samples far outside [-1, 1] neither overflow nor underflow.
func TestLevelRange(t *testing.T) {
	require.InDelta(t, 1e200, feature.RMS([]float64{1e200, -1e200}), 1e186)
	require.InDelta(t, 1e-170, feature.RMS([]float64{1e-170, -1e-170}), 1e-184)
	require.InDelta(t, 1, feature.Crest([]float64{1e-170, -1e-170}), 1e-12, "tiny is not silent")
	require.InDelta(t, 1, feature.Crest([]float64{1e200, -1e200}), 1e-12)
	require.True(t, math.IsNaN(feature.RMS([]float64{math.NaN(), 1})))
	require.True(t, math.IsInf(feature.RMS([]float64{math.Inf(1), 1}), 1))
	require.True(t, math.IsNaN(feature.Crest([]float64{math.Inf(-1), 1})))
}
