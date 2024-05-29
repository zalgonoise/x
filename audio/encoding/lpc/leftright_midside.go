package lpc

type SignalType interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~int8 | ~int16 | ~int32 | ~int64 |
		~int | ~uint |
		~float32 | ~float64
}

// MidSide returns a mid-side representation of the input left-right values.
//
// ref:  https://en.wikipedia.org/wiki/Joint_encoding
func MidSide[T SignalType](left, right T) (mid, side T) {
	mid = left + right
	side = left - right

	return mid, side
}

// LeftRight returns a left-right representation of the input mid-side values.
//
// ref:  https://en.wikipedia.org/wiki/Joint_encoding
func LeftRight[T SignalType](mid, side T) (left, right T) {
	left = (side + mid) / 2
	right = (mid - side) / 2

	return left, right
}
