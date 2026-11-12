package testutil

import (
	"math"
	"math/cmplx"
	"testing"
)

// ULPDiff returns the distance between a and b in units in the last place. Equal values (including +0 and -0)
// return 0, and a NaN on either side returns math.MaxUint64.
func ULPDiff(a, b float64) uint64 {
	switch {
	case math.IsNaN(a) || math.IsNaN(b):
		return math.MaxUint64
	case a == b:
		return 0
	}

	ia, ib := orderedBits(a), orderedBits(b)
	if ia > ib {
		return uint64(ia - ib)
	}

	return uint64(ib - ia)
}

// orderedBits maps a float64 onto an int64 so that integer order matches float order.
func orderedBits(f float64) int64 {
	b := int64(math.Float64bits(f))
	if b < 0 {
		return math.MinInt64 - b
	}

	return b
}

// RequireULPs fails the test unless want and got have the same length and every pair is within maxULPs.
func RequireULPs(tb testing.TB, want, got []float64, maxULPs uint64) {
	tb.Helper()

	if len(want) != len(got) {
		tb.Fatalf("length mismatch: want %d, got %d", len(want), len(got))
	}

	for i := range want {
		if d := ULPDiff(want[i], got[i]); d > maxULPs {
			tb.Fatalf("index %d: want %v, got %v (%d ULPs apart, max %d)", i, want[i], got[i], d, maxULPs)
		}
	}
}

// RequireInDelta fails the test unless want and got have the same length and every pair differs by at most delta.
func RequireInDelta(tb testing.TB, want, got []float64, delta float64) {
	tb.Helper()

	if len(want) != len(got) {
		tb.Fatalf("length mismatch: want %d, got %d", len(want), len(got))
	}

	for i := range want {
		if math.Abs(want[i]-got[i]) > delta || math.IsNaN(got[i]) != math.IsNaN(want[i]) {
			tb.Fatalf("index %d: want %v, got %v (delta %g)", i, want[i], got[i], delta)
		}
	}
}

// DB converts a linear amplitude ratio to decibels (20·log10). Zero returns -Inf.
func DB(ratio float64) float64 {
	return 20 * math.Log10(math.Abs(ratio))
}

// RequireDB fails the test unless got has the same sign as want and is within tolDB decibels of it.
func RequireDB(tb testing.TB, want, got, tolDB float64) {
	tb.Helper()

	if want == 0 || got == 0 {
		if want != got {
			tb.Fatalf("want %v, got %v", want, got)
		}

		return
	}

	if math.Signbit(want) != math.Signbit(got) {
		tb.Fatalf("want %v, got %v (opposite signs)", want, got)
	}

	if d := math.Abs(DB(got / want)); d > tolDB || math.IsNaN(d) {
		tb.Fatalf("want %v, got %v (%.4f dB apart, max %.4f dB)", want, got, d, tolDB)
	}
}

// NaiveDFT computes the discrete Fourier transform by its definition, X[k] = Σ x[n]·e^(-2πikn/N). It is O(N²) and
// meant as an oracle for FFT implementations on small inputs.
func NaiveDFT(x []complex128) []complex128 {
	n := len(x)
	out := make([]complex128, n)

	for k := range n {
		var sum complex128

		for t := range n {
			angle := -2 * math.Pi * float64(k) * float64(t) / float64(n)
			sum += x[t] * cmplx.Rect(1, angle)
		}

		out[k] = sum
	}

	return out
}

// Real converts real samples to complex numbers with a zero imaginary part.
func Real(x []float64) []complex128 {
	out := make([]complex128, len(x))

	for i, v := range x {
		out[i] = complex(v, 0)
	}

	return out
}

// Magnitudes returns |X[k]| for every bin of a spectrum.
func Magnitudes(spectrum []complex128) []float64 {
	out := make([]float64, len(spectrum))

	for i, v := range spectrum {
		out[i] = cmplx.Abs(v)
	}

	return out
}
