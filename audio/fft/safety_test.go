package fft_test

import (
	"fmt"
	"math"
	"slices"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/audio/fft"
	"github.com/zalgonoise/x/audio/fft/window"
	"github.com/zalgonoise/x/audio/internal/testutil"
)

func TestFFTMatchesNaiveDFT(t *testing.T) {
	for n := 4; n <= 1024; n <<= 1 {
		x := testutil.Real(testutil.Noise(n, 1, uint64(n)))
		want := testutil.NaiveDFT(x)
		got := fft.FFT(slices.Clone(x))

		require.Len(t, got, n)

		for k := range want {
			require.InDelta(t, real(want[k]), real(got[k]), 1e-9*float64(n), "n=%d bin %d (real)", n, k)
			require.InDelta(t, imag(want[k]), imag(got[k]), 1e-9*float64(n), "n=%d bin %d (imag)", n, k)
		}
	}
}

// TestFFTLargeSizesClosedForm checks the lazily computed twiddle tables (sizes beyond the precomputed 2048) against
// closed-form spectra: a bin-centred cosine of amplitude a has |X[k]| = |X[N-k]| = a*N/2 and ~0 elsewhere.
func TestFFTLargeSizesClosedForm(t *testing.T) {
	for _, n := range []int{2048, 4096, 8192} {
		const k = 37

		mags := testutil.Magnitudes(fft.FFT(testutil.Real(testutil.BinCosine(n, k, 2))))

		for i, m := range mags {
			want := 0.0
			if i == k || i == n-k {
				want = float64(n)
			}

			require.InDelta(t, want, m, 1e-6*float64(n), "n=%d bin %d", n, i)
		}
	}
}

// TestFFTConcurrentNewSizes runs FFTs of sizes beyond the precomputed twiddle tables from many goroutines at once;
// the lazily-filled cache used to be written without synchronisation (a fatal "concurrent map writes" under load).
// The input is a bin-centred cosine, whose spectrum depends on every twiddle factor.
func TestFFTConcurrentNewSizes(t *testing.T) {
	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []string
	)

	for _, n := range []int{1 << 13, 1 << 14, 1 << 15} {
		for g := range 8 {
			wg.Go(func() {
				k := 5 + g
				mags := testutil.Magnitudes(fft.FFT(testutil.Real(testutil.BinCosine(n, k, 1))))

				if math.Abs(mags[k]-float64(n)/2) > 1e-6*float64(n) || mags[k+1] > 1e-6*float64(n) {
					mu.Lock()

					errs = append(errs, fmt.Sprintf("n=%d k=%d: |X[k]|=%g |X[k+1]|=%g", n, k, mags[k], mags[k+1]))
					mu.Unlock()
				}
			})
		}
	}

	wg.Wait()
	require.Empty(t, errs)
}

func TestGetRadix2FactorsReturnsCopy(t *testing.T) {
	factors := fft.GetRadix2Factors(64)
	for i := range factors {
		factors[i] = 0 // a caller scribbling on the result must not affect later FFTs
	}

	x := testutil.Real(testutil.BinCosine(64, 5, 1))
	mags := testutil.Magnitudes(fft.FFT(x))
	require.InDelta(t, 32.0, mags[5], 1e-9)
}

func TestToComplexCapacity(t *testing.T) {
	c := fft.ToComplex([]float64{1, 2, 3})
	require.Equal(t, []complex128{1, 2, 3}, c)
	require.Equal(t, len(c), cap(c), "appending must not write past the allocation")
}

func TestApplyDoesNotMutateInput(t *testing.T) {
	data := testutil.Noise(1024, 1, 3)
	original := slices.Clone(data)

	spectrum := fft.Apply(44100, data, window.New(window.Blackman, 1024))
	require.NotEmpty(t, spectrum)
	require.Equal(t, original, data, "windowing happens on a copy")
}

func TestApplyInvalidLengths(t *testing.T) {
	for _, n := range []int{0, 1, 2, 3, 100, 1000} {
		require.Nil(t, fft.Apply(44100, make([]float64, n), nil), "length %d", n)
	}
}
