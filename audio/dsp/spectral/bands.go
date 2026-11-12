package spectral

import (
	"errors"
	"fmt"
	"math"
	"strconv"
)

// ErrInvalidBands is returned for band layouts that cannot be built.
var ErrInvalidBands = errors.New("spectral: invalid bands")

// Fractions of an octave that OctaveBands lays out.
const (
	Octave      = 1
	ThirdOctave = 3
)

// Band is the frequency range [Low, High) Hz, named after its Centre.
type Band struct {
	Low, Centre, High float64
}

// r10 holds the nominal mid-band frequencies of the third-octave bands within one decade (IEC 61260-1, the R10 series),
// times 100: integers, which nominal turns into correctly rounded frequencies (0.16, not 0.16000000000000003).
//
//nolint:gochecknoglobals // immutable table of the standard's preferred numbers
var r10 = [10]int{100, 125, 160, 200, 250, 315, 400, 500, 630, 800}

// OctaveBands returns the octave (fraction Octave) or third-octave (fraction ThirdOctave) bands of IEC 61260-1, base 10,
// whose nominal centres lie in [lo, hi] Hz, in ascending order. Centres are the nominal frequencies (31.5, 63, 125...,
// 1000...) and edges the exact ones: the exact mid-band frequency times 10^(±3/20) for octaves and 10^(±1/20) for third
// octaves, so adjacent bands share their edges.
func OctaveBands(fraction int, lo, hi float64) ([]Band, error) {
	if fraction != Octave && fraction != ThirdOctave {
		return nil, fmt.Errorf("%w: fraction 1/%d (want 1 or 3)", ErrInvalidBands, fraction)
	}

	if !(lo > 0) || !(hi >= lo) || math.IsInf(hi, 0) {
		return nil, fmt.Errorf("%w: range [%g, %g] Hz", ErrInvalidBands, lo, hi)
	}

	const (
		reference         = 1000.0 // Hz, the mid-band frequency of band 0
		referenceExponent = 3      // reference = 10³ Hz
		r10Exponent       = 2      // r10 holds the R10 values times 10²
		perDecade         = 10     // third octaves in a decade
	)

	step := ThirdOctave / fraction // third octaves between two band centres
	half := float64(step) / 2      //nolint:mnd // third octaves from the mid-band frequency to an edge

	var bands []Band

	// third-octave index x: the exact mid-band frequency is 1000·10^(x/10) Hz; start a decade below lo to round safely
	// (log10(lo) - 3 rather than log10(lo/1000), which underflows for the smallest lo)
	x := int(math.Floor(perDecade*(math.Log10(lo)-referenceExponent))) - perDecade
	x -= mod(x, step)

	for ; ; x += step {
		e := referenceExponent - r10Exponent + floorDiv(x, perDecade)
		nominal := scale10(r10[mod(x, perDecade)], e)

		if nominal > hi {
			return bands, nil
		}

		if nominal < lo {
			continue
		}

		// both edges from the same exponents as the neighbours', so that adjacent bands share them exactly
		bands = append(bands, Band{
			Low:    reference * math.Pow(perDecade, (float64(x)-half)/perDecade),
			Centre: nominal,
			High:   reference * math.Pow(perDecade, (float64(x)+half)/perDecade),
		})
	}
}

// mod returns x modulo m in [0, m), for negative x too.
func mod(x, m int) int {
	return ((x % m) + m) % m
}

// floorDiv returns ⌊x/m⌋, for negative x too.
func floorDiv(x, m int) int {
	return (x - mod(x, m)) / m
}

// scale10 returns v·10^e, correctly rounded at every scale: the decimal is parsed rather than computed, as a power of
// ten below 10^-22 or above 10^22 is inexact in float64. It overflows to +Inf and underflows to 0.
func scale10(v, e int) float64 {
	f, _ := strconv.ParseFloat(strconv.Itoa(v)+"e"+strconv.Itoa(e), 64) //nolint:errcheck // ±Inf or 0 out of range

	return f
}

// Banding maps the bins of a spectrum onto bands and reduces each band to one level. A bin belongs to the band that
// contains its centre frequency; bands that contain no bin at the spectrum's resolution are dropped.
type Banding struct {
	bands       []Band
	first, last []int // bins [first[i], last[i]) of bands[i]
	bins        int   // Size/2+1: the length of the spectra it reduces
	enbw        float64
}

// NewBanding returns the Banding of the Size/2+1 bins of a size-point spectrum of audio sampled at rate Hz, analysed
// with a window of enbw bins of equivalent noise bandwidth (STFT.ENBW). Bands must be ascending and must not overlap.
func NewBanding(bands []Band, size, rate int, enbw float64) (*Banding, error) {
	if size < 2 || rate < 1 || !(enbw > 0) || math.IsInf(enbw, 0) {
		return nil, fmt.Errorf("%w: size %d, rate %d Hz, ENBW %g", ErrInvalidBands, size, rate, enbw)
	}

	k, bins := 0, size/2+1 //nolint:mnd // a one-sided spectrum: 0 Hz to Nyquist
	b := &Banding{bins: bins, enbw: enbw}

	for i, band := range bands {
		if !(band.Low < band.High) || (i > 0 && band.Low < bands[i-1].High) {
			return nil, fmt.Errorf("%w: band %d [%g, %g) Hz is empty, or not above the previous band", ErrInvalidBands, i,
				band.Low, band.High)
		}

		for k < bins && BinFrequency(k, size, rate) < band.Low {
			k++
		}

		first := k

		for k < bins && BinFrequency(k, size, rate) < band.High {
			k++
		}

		if k > first {
			b.bands = append(b.bands, band)
			b.first = append(b.first, first)
			b.last = append(b.last, k)
		}
	}

	return b, nil
}

// Bands returns the bands that hold at least one bin, in ascending order. The slice must not be modified.
func (b *Banding) Bands() []Band { return b.bands }

// Levels appends the level of every band of Bands to dst and returns it: sqrt(Σ A[k]² / ENBW) over the band's bins of
// the amplitude spectrum. That is the amplitude of a sinusoid inside the band when the band holds the window's whole main
// lobe (and the RMS level of the band, times √2, for noise). A band narrower than the main lobe reads less, the rest
// landing in its neighbours: a bin-centred sine in a one-bin band reads 1/√ENBW of its amplitude (-1.8 dB with Hann). A
// spectrum of any length but Size/2+1 appends nothing.
func (b *Banding) Levels(spectrum, dst []float64) []float64 {
	if len(spectrum) != b.bins {
		return dst
	}

	for i := range b.bands {
		dst = append(dst, b.level(spectrum[b.first[i]:b.last[i]]))
	}

	return dst
}

// level returns sqrt(Σ a² / ENBW), scaling by the largest value first so that no square overflows or underflows.
func (b *Banding) level(a []float64) float64 {
	var peak float64

	for _, v := range a {
		peak = max(peak, math.Abs(v))
	}

	if peak == 0 || math.IsInf(peak, 0) || math.IsNaN(peak) {
		return peak
	}

	var sum float64

	for _, v := range a {
		s := v / peak
		sum += s * s
	}

	return peak * math.Sqrt(sum/b.enbw)
}
