package data

// BitDepthTypes is a type constraint joining all the different
// data types used by the supported bit depths.
type BitDepthTypes interface {
	int8 | int16 | int32 | uint32 | byte | int | float64 | float32
}

const (
	size8 = iota + 1
	size16
	size24
	size32
)

func to[F, T BitDepthTypes](from []F) []T {
	out := make([]T, len(from))

	for i := range from {
		out[i] = T(from[i])
	}

	return out
}
