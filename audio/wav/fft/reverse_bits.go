package fft

// RevBitsTypes is a type constraint interface to scope the allowed types to
// call the ReverseBits function
type RevBitsTypes interface {
	uint | uint16 | uint32 | uint64 | int | int16 | int32 | int64
}

// ReverseFirstBits returns the first `size` bits of `value` in reverse order
// from: http://graphics.stanford.edu/~seander/bithacks.html#BitReverseObvious
func ReverseFirstBits[T RevBitsTypes](value, size T) (reverse T) {
	for ; value != 0; value, size = value>>1, size-1 {
		reverse <<= 1        // left shift
		reverse |= value & 1 // add next bit
	}

	return reverse << size
}

// ReverseBits returns the bits of `value` in reverse order
// from: http://graphics.stanford.edu/~seander/bithacks.html#BitReverseObvious
func ReverseBits[T RevBitsTypes](value T) (reverse T) {
	for ; value != 0; value >>= 1 {
		reverse <<= 1        // left shift
		reverse |= value & 1 // add next bit
	}

	return reverse
}
