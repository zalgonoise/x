package window

import "github.com/zalgonoise/x/audio/trig"

var blackmanMap = map[int]Window{
	8:    Blackman8,
	16:   Blackman16,
	32:   Blackman32,
	64:   Blackman64,
	128:  Blackman128,
	256:  Blackman256,
	512:  Blackman512,
	1024: Blackman1024,
	2048: Blackman2048,
	4096: Blackman4096,
	8192: Blackman8192,
}

// Blackman returns an L-point Blackman window
// Reference: http://www.mathworks.com/help/signal/ref/blackman.html
func Blackman(i int) Window {
	w, ok := blackmanMap[i]
	if !ok {
		return newBlackman(i)
	}
	return w
}

func newBlackman(i int) Window {
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
				term1 = -0.5 * trig.Cos(tau*n/indices)
				term2 = 0.08 * trig.Cos(twoTau*n/indices)
			)

			r[int(n)] = term0 + term1 + term2
		}

		return r
	}
}
