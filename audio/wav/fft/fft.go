package fft

import (
	"math"
	"runtime"
	"sync"

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
		ln         = len(data)
		magnitudes = make([]FrequencyPower, 0, (ln/2)-1)
	)

	// apply a window function to the values
	if w != nil && len(w) == ln {
		w.Apply(data)
	}

	// apply a fast Fourier transform on the data; exclude index 0, no 0Hz-freq results
	spectrum := FFT(ToComplex(data))

	for i := 1; i < ln/2; i++ {
		freqReal := real(spectrum[i])
		freqImag := imag(spectrum[i])
		// map the magnitude for each frequency bin to the corresponding value in the map
		// using math.Sqrt(re*re + im*im) is faster than using math.Hypot(re, im)
		// see fft_test.go for details
		magnitudes = append(
			magnitudes,
			FrequencyPower{
				Freq: i * sampleRate / ln,
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

	temp := make([]complex128, lx) // temp
	reorder := reorderData(x)
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
						ridx := reorder[idx]
						w_n := reorder[idx2] * factors[blocks*j]
						temp[idx] = ridx + w_n
						temp[idx2] = ridx - w_n
					}
				} else {
					n1 := nb + 1
					rn := reorder[nb]
					rn1 := reorder[n1]
					temp[nb] = rn + rn1
					temp[n1] = rn - rn1
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
		reorder, temp = temp, reorder
	}

	return reorder
}

// TODO: refactor
func reorderData(x []complex128) []complex128 {
	ln := uint(len(x))
	reorder := make([]complex128, ln)
	s := Log2(ln)

	var n uint
	for ; n < ln; n++ {
		reorder[ReverseFirstBits(n, s)] = x[n]
	}

	return reorder
}
