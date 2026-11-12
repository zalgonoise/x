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

// frames collects every frame a Framer emits for x written in chunks of the given sizes (the rest in one call).
func frames(t *testing.T, size, hop int, x []float64, chunks ...int) [][]float64 {
	t.Helper()

	f, err := spectral.NewFramer(size, hop)
	require.NoError(t, err)

	var out [][]float64

	collect := func(frame []float64) {
		require.Len(t, frame, size)

		out = append(out, slices.Clone(frame))
	}

	for _, n := range chunks {
		n = min(n, len(x))
		f.Write(x[:n], collect)
		x = x[n:]
	}

	f.Write(x, collect)

	return out
}

// ramp returns 0, 1, ..., n-1: every frame then names the index of its first sample.
func ramp(n int) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = float64(i)
	}

	return out
}

func TestFramer(t *testing.T) {
	for _, tc := range []struct {
		name      string
		size, hop int
		n         int
		starts    []int
	}{
		{name: "half overlap", size: 4, hop: 2, n: 11, starts: []int{0, 2, 4, 6}},
		{name: "no overlap", size: 4, hop: 4, n: 11, starts: []int{0, 4}},
		{name: "skip between frames", size: 3, hop: 5, n: 14, starts: []int{0, 5, 10}},
		{name: "hop of one", size: 3, hop: 1, n: 5, starts: []int{0, 1, 2}},
		{name: "shorter than a frame", size: 8, hop: 2, n: 7},
		{name: "exactly one frame", size: 8, hop: 8, n: 8, starts: []int{0}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			x := ramp(tc.n)
			want := make([][]float64, 0, len(tc.starts))

			for _, s := range tc.starts {
				want = append(want, x[s:s+tc.size])
			}

			require.Equal(t, want, nilIfEmpty(frames(t, tc.size, tc.hop, x)), "one write")

			// the carry makes the frames independent of how the stream is split
			for split := range tc.n + 1 {
				require.Equal(t, want, nilIfEmpty(frames(t, tc.size, tc.hop, x, split)), "split at %d", split)
			}

			ones := make([]int, tc.n)
			for i := range ones {
				ones[i] = 1
			}

			require.Equal(t, want, nilIfEmpty(frames(t, tc.size, tc.hop, x, ones...)), "one sample per write")
			require.Equal(t, ramp(tc.n), x, "the input is never modified")
		})
	}
}

func nilIfEmpty(f [][]float64) [][]float64 {
	if len(f) == 0 {
		return [][]float64{}
	}

	return f
}

func TestFramerReset(t *testing.T) {
	f, err := spectral.NewFramer(4, 2)
	require.NoError(t, err)
	require.Equal(t, 4, f.Size())
	require.Equal(t, 2, f.Hop())

	var got [][]float64

	collect := func(frame []float64) { got = append(got, slices.Clone(frame)) }

	f.Write([]float64{1, 2, 3}, collect)
	f.Reset()
	f.Write([]float64{10, 11, 12, 13, 14}, collect)

	require.Equal(t, [][]float64{{10, 11, 12, 13}}, got, "no frame straddles the reset")

	// a reset also forgets samples still to skip
	f, err = spectral.NewFramer(2, 5)
	require.NoError(t, err)

	got = nil

	f.Write([]float64{1, 2, 3}, collect) // frame {1, 2}, then 3 more samples to skip
	f.Reset()
	f.Write([]float64{7, 8}, collect)

	require.Equal(t, [][]float64{{1, 2}, {7, 8}}, got)
}

func TestNewFramerInvalid(t *testing.T) {
	for _, tc := range [][2]int{{0, 1}, {1, 0}, {-1, 1}, {4, -2}} {
		_, err := spectral.NewFramer(tc[0], tc[1])
		require.ErrorIs(t, err, spectral.ErrInvalidSize, "size %d, hop %d", tc[0], tc[1])
	}
}

func TestNewSTFTInvalid(t *testing.T) {
	for _, tc := range []struct {
		name      string
		size, hop int
		window    []float64
	}{
		{name: "not a power of two", size: 1000, hop: 500},
		{name: "too small", size: 2, hop: 1},
		{name: "too large", size: spectral.MaxSize * 2, hop: 1},
		{name: "zero hop", size: 8, hop: 0},
		{name: "window length", size: 8, hop: 8, window: window.Hann(16)},
		{name: "zero window", size: 4, hop: 4, window: []float64{0, 0, 0, 0}},
		{name: "negative window", size: 4, hop: 4, window: []float64{-1, -1, -1, -1}},
		{name: "NaN window", size: 4, hop: 4, window: []float64{1, math.NaN(), 1, 1}},
		{name: "infinite window", size: 4, hop: 4, window: []float64{1, math.Inf(1), 1, 1}},
		{name: "overflowing gain", size: 4, hop: 4, window: []float64{1e308, 1e308, 1e308, 1e308}},
		{name: "underflowing gain", size: 4, hop: 4, window: []float64{5e-324, 5e-324, 0, 0}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := spectral.NewSTFT(tc.size, tc.hop, tc.window)
			require.ErrorIs(t, err, spectral.ErrInvalidSize)
		})
	}
}

// windows are the fft/window tables the tests run on; nil is rectangular.
var windows = map[string]func(int) window.Window{
	"rectangular": nil,
	"hann":        window.Hann,
	"hamming":     window.Hamming,
	"blackman":    window.Blackman,
	"bartlett":    window.Bartlett,
	"flattop":     window.FlatTop,
}

func newSTFT(t *testing.T, size, hop int, w func(int) window.Window) *spectral.STFT {
	t.Helper()

	var table []float64
	if w != nil {
		table = w(size)
	}

	s, err := spectral.NewSTFT(size, hop, table)
	require.NoError(t, err)

	return s
}

// A sine centred on bin k reads its amplitude on bin k, whatever the window; with a rectangular window every other bin
// is zero.
func TestSpectrumSineOnBin(t *testing.T) {
	const (
		size = 1024
		rate = 48000
		amp  = 0.5
	)

	for name, w := range windows {
		t.Run(name, func(t *testing.T) {
			s := newSTFT(t, size, size, w)
			require.Equal(t, size/2+1, s.Bins())

			// a window leaks the sine's mirror image (at -k) into bin k: exact only well away from DC and Nyquist
			bins, tolerance := []int{7, 100, 256, size/2 - 12}, 1e-5
			if w == nil {
				bins, tolerance = []int{1, 7, 100, 256, size/2 - 1}, 1e-9
			}

			for _, k := range bins {
				hz := spectral.BinFrequency(k, size, rate)
				x := testutil.Sine(size, rate, hz, amp, 0.3)

				got := s.Spectrum(x)
				require.Len(t, got, size/2+1)
				require.InDelta(t, amp, got[k], tolerance, "bin %d (%g Hz)", k, hz)

				if w == nil {
					for j, v := range got {
						if j != k {
							require.InDelta(t, 0, v, 1e-9, "leakage on bin %d", j)
						}
					}
				}

				// the strongest bin is the sine's
				require.Equal(t, k, argmax(got), "bin %d", k)
			}
		})
	}
}

func argmax(x []float64) int {
	best := 0

	for i, v := range x {
		if v > x[best] {
			best = i
		}
	}

	return best
}

func TestSpectrumDCAndNyquist(t *testing.T) {
	s := newSTFT(t, 64, 64, nil)

	got := s.Spectrum(testutil.DC(64, -0.25))
	require.InDelta(t, 0.25, got[0], 1e-12, "DC reads its magnitude, unscaled")

	got = s.Spectrum(testutil.Alternating(64))
	require.InDelta(t, 1, got[32], 1e-12, "the alternating component reads its amplitude on the Nyquist bin")
	require.InDelta(t, 0, got[0], 1e-12)
}

// Parseval, for the one-sided amplitude spectrum of a rectangular window: the mean square of a frame is
// A[0]² + A[N/2]² + Σ A[k]²/2 over the other bins.
func TestSpectrumParseval(t *testing.T) {
	for _, size := range []int{4, 64, 1024, 8192} {
		x := testutil.Noise(size, 0.7, uint64(size))
		s := newSTFT(t, size, size, nil)
		a := s.Spectrum(x)

		var meanSquare float64
		for _, v := range x {
			meanSquare += v * v
		}

		meanSquare /= float64(size)

		power := a[0]*a[0] + a[size/2]*a[size/2]
		for _, v := range a[1 : size/2] {
			power += v * v / 2
		}

		require.InDelta(t, meanSquare, power, 1e-12*meanSquare*float64(size), "size %d", size)
	}
}

// The FFT-based spectrum equals the scaled magnitudes of the naive DFT of the windowed frame.
func TestSpectrumMatchesNaiveDFT(t *testing.T) {
	for name, w := range windows {
		for _, size := range []int{8, 64, 256} {
			x := testutil.Noise(size, 1, 7)
			s := newSTFT(t, size, size, w)

			windowed := slices.Clone(x)
			sum := float64(size)

			if w != nil {
				table := w(size)
				sum = 0

				for i := range windowed {
					windowed[i] *= table[i]
					sum += table[i]
				}
			}

			want := testutil.Magnitudes(testutil.NaiveDFT(testutil.Real(windowed)))[:size/2+1]
			for k := range want {
				want[k] /= sum
				if k != 0 && k != size/2 {
					want[k] *= 2
				}
			}

			got := s.Spectrum(x)
			testutil.RequireInDelta(t, want, got, 1e-9)
			require.Equal(t, testutil.Noise(size, 1, 7), x, "%s/%d: the frame is never modified", name, size)
		}
	}
}

func TestSpectrumWrongLength(t *testing.T) {
	s := newSTFT(t, 16, 16, nil)
	require.Nil(t, s.Spectrum(make([]float64, 15)))
	require.Nil(t, s.Spectrum(nil))
}

// The window table the STFT was built from is never written to, and later changes to it do not reach the STFT.
func TestSTFTOwnsWindow(t *testing.T) {
	table := slices.Clone(window.Hann(64))
	s, err := spectral.NewSTFT(64, 32, table)
	require.NoError(t, err)

	x := testutil.Noise(64, 1, 3)
	before := slices.Clone(s.Spectrum(x))

	require.Equal(t, window.Hann(64), table, "untouched")

	for i := range table {
		table[i] = 0
	}

	require.Equal(t, before, s.Spectrum(x), "the STFT keeps its own copy")
}

func TestSTFTWriteCarriesAcrossCalls(t *testing.T) {
	const (
		size = 256
		hop  = 64
		rate = 8000
	)

	x := testutil.Chirp(5000, rate, 100, 3000, 0.8)

	collect := func(s *spectral.STFT, chunks ...int) [][]float64 {
		var out [][]float64

		in := x
		for _, n := range chunks {
			s.Write(in[:n], func(spectrum []float64) { out = append(out, slices.Clone(spectrum)) })
			in = in[n:]
		}

		s.Write(in, func(spectrum []float64) { out = append(out, slices.Clone(spectrum)) })

		return out
	}

	want := collect(newSTFT(t, size, hop, window.Hann))
	require.Len(t, want, (len(x)-size)/hop+1)

	got := collect(newSTFT(t, size, hop, window.Hann), 1, 100, 255, 333, 7)
	require.Equal(t, want, got)

	// every spectrum is the spectrum of its frame
	s := newSTFT(t, size, hop, window.Hann)
	for i, spectrum := range want {
		require.Equal(t, spectrum, s.Spectrum(x[i*hop:i*hop+size]), "frame %d", i)
	}

	s.Reset()

	var n int

	s.Write(x[:size-1], func([]float64) { n++ })
	s.Reset()
	s.Write(x[:size-1], func([]float64) { n++ })
	require.Zero(t, n, "reset drops the carried samples")
}

func TestENBW(t *testing.T) {
	require.InDelta(t, 1, newSTFT(t, 1024, 1024, nil).ENBW(), 1e-12)
	// the symmetric Hann window's ENBW is 1.5·N/(N-1)
	require.InDelta(t, 1.5*1024/1023, newSTFT(t, 1024, 1024, window.Hann).ENBW(), 1e-6)

	s := newSTFT(t, 1024, 512, window.Hann)
	require.Equal(t, 1024, s.Size())
	require.Equal(t, 512, s.Hop())
}

func TestBinFrequency(t *testing.T) {
	require.InDelta(t, 2002.587890625, spectral.BinFrequency(93, 2048, 44100), 0, "float Hz, not truncated")
	require.InDelta(t, 22050, spectral.BinFrequency(1024, 2048, 44100), 0)
	require.Zero(t, spectral.BinFrequency(0, 2048, 44100))
}

// A clone carries on from the same samples, independently of the original, and shares the window without changing it.
func TestSTFTClone(t *testing.T) {
	const (
		size = 64
		hop  = 16
	)

	x := testutil.Noise(400, 1, 9)
	table := slices.Clone(window.Hann(size))

	collect := func(s *spectral.STFT, x []float64) [][]float64 {
		var out [][]float64

		s.Write(x, func(spectrum []float64) { out = append(out, slices.Clone(spectrum)) })

		return out
	}

	template, err := spectral.NewSTFT(size, hop, table)
	require.NoError(t, err)

	a := template.Clone()
	b := template.Clone()

	want := collect(newSTFT(t, size, hop, window.Hann), x)
	require.Equal(t, want, collect(a, x), "a clone of a fresh STFT starts empty")
	require.Equal(t, want, collect(b, x), "clones do not share their carried samples")
	require.Empty(t, collect(template, x[:size-1]), "nor do they feed the template")

	// a clone taken mid-stream carries on from the same samples
	s := newSTFT(t, size, hop, window.Hann)
	head := collect(s, x[:150])
	c := s.Clone()

	tail := collect(s, x[150:])
	require.Equal(t, want, append(head, tail...))
	require.Equal(t, tail, collect(c, x[150:]))
	require.Equal(t, window.Hann(size), table, "the window is never written")

	require.Equal(t, size/2+1, a.Bins())
	require.Equal(t, template.ENBW(), a.ENBW())
}
