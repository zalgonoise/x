package spectral_test

import (
	"math"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/dsp/spectral"
	"github.com/zalgonoise/x/audio/fft/window"
	"github.com/zalgonoise/x/audio/internal/testutil"
)

// Review of the spectrum code (2026-11-02).

// A Banding reduces spectra of exactly Size/2+1 bins: one of another frame size has its bins at other frequencies.
func TestLevelsRejectsOtherSpectrumLengths(t *testing.T) {
	bands, err := spectral.OctaveBands(spectral.Octave, 63, 4000)
	require.NoError(t, err)

	b, err := spectral.NewBanding(bands, 4096, 48000, 1)
	require.NoError(t, err)
	require.NotEmpty(t, b.Bands())

	for _, n := range []int{2048/2 + 1, 8192/2 + 1, 0} {
		spectrum := make([]float64, n)
		if n > 100 {
			spectrum[100] = 1
		}

		require.Empty(t, b.Levels(spectrum, nil), "%d bins", n)
	}

	require.Len(t, b.Levels(make([]float64, 4096/2+1), nil), len(b.Bands()))
}

// A band narrower than the window's main lobe reads 1/√ENBW of a bin-centred sine, as documented: the 50 Hz third
// octave holds one bin at 48 kHz and 4096 points.
func TestBandLevelInNarrowBand(t *testing.T) {
	const (
		size = 4096
		rate = 48000
		amp  = 0.25
	)

	s, err := spectral.NewSTFT(size, size, window.Hann(size))
	require.NoError(t, err)

	bands, err := spectral.OctaveBands(spectral.ThirdOctave, 20, 20000)
	require.NoError(t, err)

	b, err := spectral.NewBanding(bands, size, rate, s.ENBW())
	require.NoError(t, err)

	i := slices.IndexFunc(b.Bands(), func(band spectral.Band) bool { return band.Centre == 50 })
	require.GreaterOrEqual(t, i, 0)

	levels := b.Levels(s.Spectrum(testutil.Sine(size, rate, spectral.BinFrequency(4, size, rate), amp, 0)), nil)
	require.InEpsilon(t, amp/math.Sqrt(s.ENBW()), levels[i], 1e-3)
}

// The ENBW does not depend on the window's scale, and neither do the amplitudes.
func TestScaledWindows(t *testing.T) {
	x := testutil.Sine(64, 64, 8, 0.5, 0)

	for _, v := range []float64{6e153, 1e-170, 1e300} {
		s, err := spectral.NewSTFT(64, 64, slices.Repeat([]float64{v}, 64))
		require.NoError(t, err, "%g", v)
		require.InDelta(t, 1, s.ENBW(), 1e-12, "a rectangular window of %g", v)
		require.InDelta(t, 0.5, s.Spectrum(x)[8], 1e-12, "a rectangular window of %g", v)
	}
}

// Nominal centres are the R10 values, correctly rounded at every scale.
func TestNominalCentresAtEveryScale(t *testing.T) {
	bands, err := spectral.OctaveBands(spectral.ThirdOctave, 0.1, 1)
	require.NoError(t, err)
	require.Equal(t, []float64{0.1, 0.125, 0.16, 0.2, 0.25, 0.315, 0.4, 0.5, 0.63, 0.8, 1}, centres(bands))

	for _, f := range []float64{0.16, 1.25e-6, 2.5e-6, 5e-6, 1.6e-5, 1e-300, 31.5, 12500, 1.6e10} {
		one, err := spectral.OctaveBands(spectral.ThirdOctave, f, f)
		require.NoError(t, err)
		require.Equal(t, []float64{f}, centres(one), "the range is inclusive at both ends")
	}
}

func TestOctaveBandsTinyLo(t *testing.T) {
	bands, err := spectral.OctaveBands(spectral.ThirdOctave, 1e-321, 1)
	require.NoError(t, err)
	require.NotEmpty(t, bands)
	require.InDelta(t, 1.0, bands[len(bands)-1].Centre, 0)

	for i := 1; i < len(bands); i++ {
		require.Greater(t, bands[i].Centre, bands[i-1].Centre)
	}
}

// Reset from the frame callback starts the next frame with the rest of the input, whatever the hop.
func TestResetInsideCallback(t *testing.T) {
	for _, tc := range []struct {
		size, hop int
		want      [][]float64
	}{
		{size: 4, hop: 2, want: [][]float64{{0, 1, 2, 3}, {4, 5, 6, 7}}},
		{size: 4, hop: 4, want: [][]float64{{0, 1, 2, 3}, {4, 5, 6, 7}}},
		{size: 2, hop: 5, want: [][]float64{{0, 1}, {2, 3}, {7, 8}}},
	} {
		f, err := spectral.NewFramer(tc.size, tc.hop)
		require.NoError(t, err)

		var got [][]float64

		require.NotPanics(t, func() {
			f.Write(ramp(9), func(frame []float64) {
				got = append(got, slices.Clone(frame))
				if len(got) == 1 {
					f.Reset()
				}
			})
		})

		require.Equal(t, tc.want, got, "size %d, hop %d", tc.size, tc.hop)
	}
}

// A Framer allocates as samples arrive: an absurd frame size costs nothing until it is fed.
func TestNewFramerHugeSize(t *testing.T) {
	require.NotPanics(t, func() {
		f, err := spectral.NewFramer(math.MaxInt, 1)
		require.NoError(t, err)

		f.Write(make([]float64, 10), func([]float64) { t.Fatal("no frame completes") })
	})
}
