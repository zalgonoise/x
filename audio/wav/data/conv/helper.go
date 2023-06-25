package conv

// BitDepthTypes is a type constraint joining all the different
// data types used by the supported bit depths
type BitDepthTypes interface {
	int8 | int16 | int32 | uint32 | byte | int | float64 | float32
}

const (
	size8 = iota + 1
	size16
	size24
	size32
)

func convert[F, T BitDepthTypes](from []F, fn func(F) T) []T {
	out := make([]T, len(from))
	for i := range from {
		out[i] = fn(from[i])
	}
	return out
}

func copy24to32(b []byte) []byte {
	var (
		ln     = len(b)
		newLen = ln + (ln / size24)
		n      int
	)

	// if byte length does not end in a full block,
	// a full block is added instead
	if ln%size24 > 0 {
		newLen += size24
	}

	out := make([]byte, newLen)
	for j := 0; n < ln; j += size32 {
		// slice index out-of-bounds protection
		if n+size24 > len(b) {
			copy(out[j:], b[n:])
			return out
		}

		n += copy(out[j:], b[n:n+size24])
	}

	return out
}

// can't inline a pointer cast and convert an array to a slice
func append2Bytes(idx int, dst []byte, src [size16]byte) {
	if idx*size16 < len(dst) {
		copy(dst[idx*size16:], src[:])
	}
}

// can't inline a pointer cast and convert an array to a slice
func append3Bytes(idx int, dst []byte, src [size24]byte) {
	if idx*size24 < len(dst) {
		copy(dst[idx*size24:], src[:])
	}
}

// can't inline a pointer cast and convert an array to a slice
func append4Bytes(idx int, dst []byte, src [size32]byte) {
	if idx*size32 < len(dst) {
		copy(dst[idx*size32:], src[:])
	}
}
