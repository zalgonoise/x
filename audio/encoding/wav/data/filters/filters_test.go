package filters_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/encoding/wav/data/filters"
)

func TestNormalize(t *testing.T) {
	buf := []float64{0.25, -0.5, 0.1}
	filters.Normalize()(buf)
	require.InDeltaSlice(t, []float64{0.5, -1, 0.2}, buf, 1e-12, "gain is 1/peak: the loudest sample reaches full scale")

	silence := []float64{0, 0, 0}
	filters.Normalize()(silence)
	require.Equal(t, []float64{0, 0, 0}, silence, "silence stays silent (no NaN)")
}

func TestAmplifyClips(t *testing.T) {
	buf := []float64{0.6, -0.6}
	filters.Amplify(2)(buf)
	require.Equal(t, []float64{1, -1}, buf)
}

func TestPhaseFlip(t *testing.T) {
	buf := []float64{0.5, -0.25}
	filters.PhaseFlip()(buf)
	require.Equal(t, []float64{-0.5, 0.25}, buf)
}

func TestAddSub(t *testing.T) {
	buf := []float64{1, 0, -1, 0.5}
	filters.Add([]float64{1, -1})(buf)
	require.Equal(t, []float64{1, -0.5, 0, -0.25}, buf, "average, repeating the shorter signal")

	buf = []float64{1, 0}
	filters.Sub([]float64{1, -1})(buf)
	require.Equal(t, []float64{0, 0.5}, buf)

	buf = []float64{0.3}
	filters.Add(nil)(buf)
	filters.Sub(nil)(buf)
	require.Equal(t, []float64{0.3}, buf, "an empty signal is a no-op")
}
