package spectral_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/dsp/spectral"
	"github.com/zalgonoise/x/audio/fft/window"
	"github.com/zalgonoise/x/audio/internal/testutil"
)

func centres(bands []spectral.Band) []float64 {
	out := make([]float64, 0, len(bands))
	for _, b := range bands {
		out = append(out, b.Centre)
	}

	return out
}

func TestOctaveBands(t *testing.T) {
	octaves, err := spectral.OctaveBands(spectral.Octave, 16, 16000)
	require.NoError(t, err)
	require.Equal(t, []float64{16, 31.5, 63, 125, 250, 500, 1000, 2000, 4000, 8000, 16000}, centres(octaves))

	thirds, err := spectral.OctaveBands(spectral.ThirdOctave, 20, 20000)
	require.NoError(t, err)
	require.Equal(t, []float64{
		20, 25, 31.5, 40, 50, 63, 80, 100, 125, 160, 200, 250, 315, 400, 500, 630, 800,
		1000, 1250, 1600, 2000, 2500, 3150, 4000, 5000, 6300, 8000, 10000, 12500, 16000, 20000,
	}, centres(thirds))

	// exact edges (IEC 61260-1, base 10): mid-band 1000·10^(x/10), edges ×10^(±1/20) or ×10^(±3/20)
	for _, tc := range []struct {
		bands []spectral.Band
		x     float64 // third-octave index of the band's mid-band frequency
		half  float64
	}{
		{bands: thirds, x: -17, half: 0.05},
		{bands: octaves, x: -18, half: 0.15},
	} {
		for i, b := range tc.bands {
			exact := 1000 * math.Pow(10, (tc.x+float64(i)*(tc.half*20))/10)
			require.InDelta(t, exact*math.Pow(10, -tc.half), b.Low, 1e-9*exact, "band %g", b.Centre)
			require.InDelta(t, exact*math.Pow(10, tc.half), b.High, 1e-9*exact, "band %g", b.Centre)
			require.InEpsilon(t, b.Centre, exact, 0.03, "the nominal centre rounds the exact one")

			if i > 0 {
				require.InDelta(t, tc.bands[i-1].High, b.Low, 1e-9*exact, "adjacent bands share their edges")
			}
		}
	}

	// the range selects nominal centres, inclusive at both ends
	some, err := spectral.OctaveBands(spectral.Octave, 63, 250)
	require.NoError(t, err)
	require.Equal(t, []float64{63, 125, 250}, centres(some))

	some, err = spectral.OctaveBands(spectral.ThirdOctave, 0.5, 1.3)
	require.NoError(t, err)
	require.Equal(t, []float64{0.5, 0.63, 0.8, 1, 1.25}, centres(some), "below 1 Hz too")

	none, err := spectral.OctaveBands(spectral.Octave, 1001, 1999)
	require.NoError(t, err)
	require.Empty(t, none)
}

func TestOctaveBandsInvalid(t *testing.T) {
	for _, tc := range []struct {
		fraction int
		lo, hi   float64
	}{
		{fraction: 2, lo: 20, hi: 20000},
		{fraction: 0, lo: 20, hi: 20000},
		{fraction: 1, lo: 0, hi: 20000},
		{fraction: 1, lo: -1, hi: 20000},
		{fraction: 1, lo: 100, hi: 50},
		{fraction: 3, lo: math.NaN(), hi: 20000},
		{fraction: 3, lo: 20, hi: math.NaN()},
		{fraction: 3, lo: 20, hi: math.Inf(1)},
	} {
		_, err := spectral.OctaveBands(tc.fraction, tc.lo, tc.hi)
		require.ErrorIs(t, err, spectral.ErrInvalidBands, "%+v", tc)
	}
}

func TestBandingMapsBins(t *testing.T) {
	// 16 points at 1600 Hz: bins every 100 Hz, from 0 to 800 Hz
	bands := []spectral.Band{
		{Low: 50, Centre: 100, High: 150},  // bin 1
		{Low: 150, Centre: 300, High: 450}, // bins 2-4
		{Low: 460, Centre: 480, High: 490}, // no bin: dropped
		{Low: 500, Centre: 650, High: 801}, // bins 5-8 (800 Hz is the Nyquist bin)
		{Low: 900, Centre: 950, High: 1000},
	}

	b, err := spectral.NewBanding(bands, 16, 1600, 1)
	require.NoError(t, err)
	require.Equal(t, []spectral.Band{bands[0], bands[1], bands[3]}, b.Bands())

	spectrum := []float64{9, 1, 1, 2, 2, 3, 0, 0, 4}
	require.InDeltaSlice(t, []float64{1, 3, 5}, b.Levels(spectrum, nil), 1e-12, "sqrt of the sum of squares")

	require.Equal(t, []float64{7}, b.Levels(spectrum[:8], []float64{7}), "too short a spectrum appends nothing")

	b, err = spectral.NewBanding(bands, 16, 1600, 4)
	require.NoError(t, err)
	require.InDeltaSlice(t, []float64{0.5, 1.5, 2.5}, b.Levels(spectrum, nil), 1e-12, "corrected for the ENBW")
}

func TestNewBandingInvalid(t *testing.T) {
	good := []spectral.Band{{Low: 100, Centre: 150, High: 200}}

	for _, tc := range []struct {
		name  string
		bands []spectral.Band
		size  int
		rate  int
		enbw  float64
	}{
		{name: "size", bands: good, size: 1, rate: 1000, enbw: 1},
		{name: "rate", bands: good, size: 16, rate: 0, enbw: 1},
		{name: "enbw", bands: good, size: 16, rate: 1000, enbw: 0},
		{name: "NaN enbw", bands: good, size: 16, rate: 1000, enbw: math.NaN()},
		{name: "empty band", bands: []spectral.Band{{Low: 100, High: 100}}, size: 16, rate: 1000, enbw: 1},
		{name: "NaN band", bands: []spectral.Band{{Low: math.NaN(), High: 100}}, size: 16, rate: 1000, enbw: 1},
		{
			name:  "overlapping bands",
			bands: []spectral.Band{{Low: 100, High: 200}, {Low: 150, High: 300}},
			size:  16, rate: 1000, enbw: 1,
		},
		{
			name:  "descending bands",
			bands: []spectral.Band{{Low: 300, High: 400}, {Low: 100, High: 200}},
			size:  16, rate: 1000, enbw: 1,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := spectral.NewBanding(tc.bands, tc.size, tc.rate, tc.enbw)
			require.ErrorIs(t, err, spectral.ErrInvalidBands)
		})
	}
}

// A sine centred on a bin inside a band reads its amplitude as the band's level, whatever the window: the window's
// leakage into the neighbouring bins is undone by the ENBW correction.
func TestBandLevelOfSine(t *testing.T) {
	const (
		size = 4096
		rate = 48000
		amp  = 0.25
	)

	bands, err := spectral.OctaveBands(spectral.Octave, 63, 16000)
	require.NoError(t, err)

	for name, w := range windows {
		if name == "flattop" {
			continue // its main lobe is 9 bins wide: the 63 Hz band does not hold it
		}

		leakage := amp / 100 // -40 dB
		if name == "bartlett" {
			leakage = amp / 20 // its side lobes decay slowly: about -40 dB a whole octave away
		}

		t.Run(name, func(t *testing.T) {
			s := newSTFT(t, size, size, w)

			b, err := spectral.NewBanding(bands, size, rate, s.ENBW())
			require.NoError(t, err)
			require.Len(t, b.Bands(), len(bands), "every octave from 63 Hz holds bins at 11.7 Hz resolution")

			for i, band := range b.Bands() {
				k := int(math.Round(band.Centre * size / rate))
				x := testutil.Sine(size, rate, spectral.BinFrequency(k, size, rate), amp, 0)

				levels := b.Levels(s.Spectrum(x), nil)
				require.InEpsilon(t, amp, levels[i], 0.02, "band %g Hz", band.Centre)

				for j, l := range levels {
					if j != i {
						require.Less(t, l, leakage, "band %g Hz: leakage into the %g Hz band", band.Centre,
							b.Bands()[j].Centre)
					}
				}
			}
		})
	}
}

// For noise, band levels are √2 times the RMS of the band: the bands of a full spectrum add up to the signal's power.
func TestBandLevelsOfNoise(t *testing.T) {
	const (
		size = 8192
		rate = 8192
	)

	x := testutil.Noise(size, 1, 11)

	var meanSquare float64
	for _, v := range x {
		meanSquare += v * v
	}

	meanSquare /= size

	s := newSTFT(t, size, size, window.Hann)

	// bands covering every bin but DC and Nyquist, which a one-sided level does not double
	bands := []spectral.Band{{Low: 0.5, High: 1000}, {Low: 1000, High: 2000}, {Low: 2000, High: rate/2 - 0.5}}

	b, err := spectral.NewBanding(bands, size, rate, s.ENBW())
	require.NoError(t, err)

	var power float64
	for _, l := range b.Levels(s.Spectrum(x), nil) {
		power += l * l / 2
	}

	// a windowed frame weighs its middle more: the estimate of uniform noise is close but not exact
	require.InEpsilon(t, meanSquare, power, 0.05)
}

func TestBandLevelExtremes(t *testing.T) {
	b, err := spectral.NewBanding([]spectral.Band{{Low: 0.5, High: 2.5}}, 4, 4, 1)
	require.NoError(t, err)

	got := b.Levels([]float64{0, 1e200, 1e200}, nil)
	require.InEpsilon(t, math.Sqrt2*1e200, got[0], 1e-12, "no overflow")

	got = b.Levels([]float64{0, 3e-170, 4e-170}, nil)
	require.InEpsilon(t, 5e-170, got[0], 1e-12, "no underflow")

	require.Equal(t, []float64{0}, b.Levels([]float64{5, 0, 0}, nil), "silence")
	require.True(t, math.IsNaN(b.Levels([]float64{0, math.NaN(), 1}, nil)[0]), "NaN propagates")
	require.True(t, math.IsInf(b.Levels([]float64{0, math.Inf(1), 1}, nil)[0], 1), "Inf propagates")
}
