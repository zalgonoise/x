package gbuf

// Repeat returns a new T item slice consisting of count copies of b.
//
// It panics if count is negative or if
// the result of (len(b) * count) overflows.
func Repeat[T any](b []T, count int) []T {
	if count == 0 {
		return []T{}
	}
	// Since we cannot return an error on overflow,
	// we should panic if the repeat will generate
	// an overflow.
	// See Issue golang.org/issue/16237.
	if count < 0 {
		panic("gbuf: negative Repeat count")
	} else if len(b)*count/count != len(b) {
		panic("gbuf: Repeat count causes overflow")
	}

	nb := make([]T, len(b)*count)
	bp := copy(nb, b)
	for bp < len(nb) {
		copy(nb[bp:], nb[:bp])
		bp *= 2
	}
	return nb
}
