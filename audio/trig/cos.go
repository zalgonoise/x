//nolint:gomnd // package contains a modified port of the standard library logic
package trig

import (
	"unsafe"

	"github.com/zalgonoise/x/ptr"
)

// Cos is an attempt to improve the performance of the math.Cos implementation.
//
// It will skip the checks for NaN and -/+Inf, so it should be perceived as an unsafe Sin / Cos approach.
// So far, I couldn't find an issue with this approach in digital signal processing; although it would be
// no longer used for audio DSP if that were to be the case.
func Cos(x float64) float64 {
	// make argument positive
	sign := false

	// raw math.Abs(x)
	x = *(*float64)(unsafe.Pointer(ptr.To(*(*uint64)(unsafe.Pointer(&x)) &^ (1 << 63))))

	var (
		j    uint64
		y, z float64
	)

	if x >= reduceThreshold {
		j, z = trigReduce(x)
	} else {
		j = uint64(x * (PI4)) // integer part of x/(Pi/4), as integer for tests on the phase angle
		y = float64(j)        // integer part of x/(Pi/4), as float

		// map zeros to origin
		if j&1 == 1 {
			j++
			y++
		}

		j &= 7                               // octant modulo 2Pi radians (360 degrees)
		z = ((x - y*PI4A) - y*PI4B) - y*PI4C // Extended precision modular arithmetic
	}

	if j > 3 {
		j -= 4
		sign = !sign
	}

	if j > 1 {
		sign = !sign
	}

	zz := z * z

	if j == 1 || j == 2 {
		y = z + z*zz*((((((_sin[0]*zz)+_sin[1])*zz+_sin[2])*zz+_sin[3])*zz+_sin[4])*zz+_sin[5])
	} else {
		y = 1.0 - 0.5*zz + zz*zz*((((((_cos[0]*zz)+_cos[1])*zz+_cos[2])*zz+_cos[3])*zz+_cos[4])*zz+_cos[5])
	}

	if sign {
		y = -y
	}

	return y
}
