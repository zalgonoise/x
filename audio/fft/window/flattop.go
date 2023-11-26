//nolint:gomnd // contains hardcoded constants, it's less readable to make constants from them
package window

import "github.com/zalgonoise/x/audio/trig"

//nolint:gochecknoglobals // immutable map linking window sizes to its corresponding precomputed window slices
var flattopMap = map[int]Window{
	8:    FlatTop8,
	16:   FlatTop16,
	32:   FlatTop32,
	64:   FlatTop64,
	128:  FlatTop128,
	256:  FlatTop256,
	512:  FlatTop512,
	1024: FlatTop1024,
	2048: FlatTop2048,
	4096: FlatTop4096,
	8192: FlatTop8192,
}

// FlatTop returns an L-point flat top window.
// Reference: http://www.mathworks.com/help/signal/ref/flattopwin.html
func FlatTop(i int) Window {
	w, ok := flattopMap[i]
	if !ok {
		return newFlatTop(i)
	}

	return w
}

func newFlatTop(i int) Window {
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
			r           = make([]float64, i)
			indices     = float64(i - 1)
			coefficient = tau / float64(i)
		)

		for n := 0.0; n <= indices; n++ {
			var (
				factor = n * coefficient
				term0  = alpha0
				term1  = alpha1 * trig.Cos(factor)
				term2  = alpha2 * trig.Cos(2*factor)
				term3  = alpha3 * trig.Cos(3*factor)
				term4  = alpha4 * trig.Cos(4*factor)
			)

			r[int(n)] = term0 - term1 + term2 - term3 + term4
		}

		return r
	}
}
