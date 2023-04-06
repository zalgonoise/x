package fft

import (
	"math"
)

type WindowFunc func(int) Block

type WindowBlock interface {
	Apply(v []float64)
	Len() int
}

type Block []float64

func (b Block) Apply(v []float64) {
	for i := range v {
		v[i] *= b[i]
	}
}

func (b Block) Len() int {
	return len(b)
}

// Hamming returns an L-point symmetric Hamming window.
// Reference: http://www.mathworks.com/help/signal/ref/hamming.html
func Hamming(i int) Block {
	switch i {
	case 0:
		return []float64{}
	case 1:
		return []float64{1}
	default:
		var (
			r           = make([]float64, i, i)
			idx         = float64(i - 1)
			coefficient = tau / idx
		)

		for n := 0.0; n <= idx; n++ {
			r[int(n)] = 0.54 - 0.46*math.Cos(coefficient*n)
		}

		return r
	}
}

// Hann returns an L-point Hann window.
// Reference: http://www.mathworks.com/help/signal/ref/hann.html
func Hann(i int) Block {
	switch i {
	case 0:
		return []float64{}
	case 1:
		return []float64{1}
	default:
		var (
			r           = make([]float64, i, i)
			indices     = float64(i - 1)
			coefficient = tau / indices
		)

		for n := 0.0; n <= indices; n++ {
			r[int(n)] = 0.5 * (1 - math.Cos(coefficient*n))
		}
		return r
	}
}

// Bartlett returns an L-point Bartlett window.
// Reference: http://www.mathworks.com/help/signal/ref/bartlett.html
func Bartlett(i int) Block {
	switch i {
	case 0:
		return []float64{}
	case 1:
		return []float64{1}
	default:
		var (
			r           = make([]float64, i, i)
			indices     = float64(i - 1)
			coefficient = 2 / indices
			n           = 0.0
		)

		for ; n <= indices/2; n++ {
			r[int(n)] = coefficient * n
		}
		for ; n <= indices; n++ {
			r[int(n)] = 2 - coefficient*n
		}

		return r
	}
}

// FlatTop returns an L-point flat top window.
// Reference: http://www.mathworks.com/help/signal/ref/flattopwin.html
func FlatTop(i int) Block {
	const (
		alpha0 = float64(0.21557895)
		alpha1 = float64(0.41663158)
		alpha2 = float64(0.277263158)
		alpha3 = float64(0.083578947)
		alpha4 = float64(0.006947368)
	)

	switch i {
	case 0:
		return []float64{}
	case 1:
		return []float64{1}
	default:
		var (
			r           = make([]float64, i, i)
			indices     = float64(i - 1)
			coefficient = tau / float64(i)
		)

		for n := 0.0; n <= indices; n++ {
			var (
				factor = n * coefficient
				term0  = alpha0
				term1  = alpha1 * math.Cos(factor)
				term2  = alpha2 * math.Cos(2*factor)
				term3  = alpha3 * math.Cos(3*factor)
				term4  = alpha4 * math.Cos(4*factor)
			)

			r[int(n)] = term0 - term1 + term2 - term3 + term4
		}

		return r
	}
}

// Blackman returns an L-point Blackman window
// Reference: http://www.mathworks.com/help/signal/ref/blackman.html
func Blackman(i int) Block {
	const term0 = 0.42

	switch i {
	case 0:
		return []float64{}
	case 1:
		return []float64{1}
	default:
		var (
			r       = make([]float64, i, i)
			indices = float64(i - 1)
		)

		for n := 0.0; n <= indices; n++ {
			var (
				term1 = -0.5 * math.Cos(tau*n/indices)
				term2 = 0.08 * math.Cos(twoTau*n/indices)
			)

			r[int(n)] = term0 + term1 + term2
		}

		return r
	}
}
