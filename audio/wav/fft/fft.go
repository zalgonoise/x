package fft

import (
	"math"
	"runtime"
	"sync"

	dspfft "github.com/mjibson/go-dsp/fft"
	"github.com/zalgonoise/x/audio/wav/fft/window"
)

var (
	radix2Factors = map[int][]complex128{
		4: {(1 + 0i), (0 - 1i), (-1 + 0i), (0 + 1i)},
	}
	worker_pool_size = 0 // TODO: remove after refactor
)

const (
	tau = 2 * math.Pi

	// DefaultMagnitudeThreshold describes the default value where a certain
	// frequency is strong enough to be considered relevant to the spectrum filter
	DefaultMagnitudeThreshold = 10
)

// FrequencyPower denotes a single frequency and its magnitude in a Fast
// Fourier Transform of a signal
type FrequencyPower struct {
	Freq int
	Mag  float64
}

// TODO: refactor / remove
type fft_work struct {
	start, end int
}

// Apply applies a Fast Fourier Transform (FFT) on a slice of float64 `data`,
// with sample rate `sampleRate`. It returns a slice of FrequencyPower
func Apply(sampleRate int, data []float64, w window.Window) []FrequencyPower {
	var (
		n          = len(data)
		freqUnit   = sampleRate / n
		magnitudes = make([]FrequencyPower, 0, (n/2)-1)
	)

	// apply a window function to the values
	if w != nil && len(w) == n {
		w.Apply(data)
	}

	// apply a fast Fourier transform on the data; exclude index 0, no 0Hz-freq results
	frequencies := dspfft.FFTReal(data)

	for i := 1; i < n/2; i++ {
		freqReal := real(frequencies[i])
		freqImag := imag(frequencies[i])
		// map the magnitude for each frequency bin to the corresponding value in the map
		// using math.Sqrt(re*re + im*im) is faster than using math.Hypot(re, im)
		// see fft_test.go for details
		magnitudes = append(
			magnitudes,
			FrequencyPower{
				Freq: i * freqUnit,
				Mag:  math.Sqrt(freqReal*freqReal + freqImag*freqImag),
			},
		)
	}
	return magnitudes
}

// TODO: continue to refactor
func FFT(x []complex128) []complex128 {
	lx := len(x)
	factors := GetRadix2Factors(lx)

	t := make([]complex128, lx) // temp
	r := reorderData(x)

	var blocks, stage, s_2 int

	jobs := make(chan *fft_work, lx)
	wg := sync.WaitGroup{}

	num_workers := worker_pool_size
	if (num_workers) == 0 {
		num_workers = runtime.GOMAXPROCS(0)
	}

	idx_diff := lx / num_workers
	if idx_diff < 2 {
		idx_diff = 2
	}

	worker := func() {
		for work := range jobs {
			for nb := work.start; nb < work.end; nb += stage {
				if stage != 2 {
					for j := 0; j < s_2; j++ {
						idx := j + nb
						idx2 := idx + s_2
						ridx := r[idx]
						w_n := r[idx2] * factors[blocks*j]
						t[idx] = ridx + w_n
						t[idx2] = ridx - w_n
					}
				} else {
					n1 := nb + 1
					rn := r[nb]
					rn1 := r[n1]
					t[nb] = rn + rn1
					t[n1] = rn - rn1
				}
			}
			wg.Done()
		}
	}

	for i := 0; i < num_workers; i++ {
		go worker()
	}
	defer close(jobs)

	for stage = 2; stage <= lx; stage <<= 1 {
		blocks = lx / stage
		s_2 = stage / 2

		for start, end := 0, stage; ; {
			if end-start >= idx_diff || end == lx {
				wg.Add(1)
				jobs <- &fft_work{start, end}

				if end == lx {
					break
				}

				start = end
			}

			end += stage
		}
		wg.Wait()
		r, t = t, r
	}

	return r
}

// TODO: refactor
func reorderData(x []complex128) []complex128 {
	lx := uint(len(x))
	r := make([]complex128, lx)
	s := log2(lx)

	var n uint
	for ; n < lx; n++ {
		r[reverseBits(n, s)] = x[n]
	}

	return r
}

// log2 returns the log base 2 of v
// from: http://graphics.stanford.edu/~seander/bithacks.html#IntegerLogObvious
// TODO: review / refactor
func log2(v uint) uint {
	var r uint

	for v >>= 1; v != 0; v >>= 1 {
		r++
	}

	return r
}

// reverseBits returns the first s bits of v in reverse order
// from: http://graphics.stanford.edu/~seander/bithacks.html#BitReverseObvious
// TODO: review / refactor
func reverseBits(v, s uint) uint {
	var r uint

	// Since we aren't reversing all the bits in v (just the first s bits),
	// we only need the first bit of v instead of a full copy.
	r = v & 1
	s--

	for v >>= 1; v != 0; v >>= 1 {
		r <<= 1
		r |= v & 1
		s--
	}

	return r << s
}

// GetRadix2Factors is temporarily public, could become private at a later point.
func GetRadix2Factors(inputLen int) []complex128 {
	if factors, ok := radix2Factors[inputLen]; ok {
		return factors
	}

	for factor, prev := 8, 4; factor <= inputLen; factor, prev = factor<<1, factor {
		if _, ok := radix2Factors[factor]; !ok {
			radix2Factors[factor] = make([]complex128, factor)

			for n, j := 0, 0; n < factor; n, j = n+2, j+1 {
				radix2Factors[factor][n] = radix2Factors[prev][j]
			}

			for n := 1; n < factor; n += 2 {
				radix2Factors[factor][n] = complex(
					math.Sincos(-tau / float64(factor) * float64(n)),
				)
			}
		}
	}

	return radix2Factors[inputLen]
}
