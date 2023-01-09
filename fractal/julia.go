package fractal

import "math/cmplx"

func Julia(z, c complex128) (n int) {
	for ; cmplx.Abs(z) < 2 && n < maxIter; n++ {
		z = z*z + c
	}
	return n
}
