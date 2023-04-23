package fft

// Log2Types is a type constraint interface to scope the allowed types to
// call the Log2 function
type Log2Types interface {
	uint | uint16 | uint32 | uint64 | int | int16 | int32 | int64
}

// Log2 returns the log base 2 of v, matching the input to precomputed
// values to achieve the fastest performance possible.
//
// If the input is not present, then its log2 value is computed with a fast
// bit-shifting technique
//
// from: http://graphics.stanford.edu/~seander/bithacks.html#IntegerLogObvious
func Log2[T Log2Types](v T) T {
	switch v {
	case 0, 1:
		return 0
	case 2:
		return 1
	case 4:
		return 2
	case 8:
		return 3
	case 16:
		return 4
	case 32:
		return 5
	case 64:
		return 6
	case 128:
		return 7
	case 256:
		return 8
	case 512:
		return 9
	case 1024:
		return 10
	case 2048:
		return 11
	case 4096:
		return 12
	case 8192:
		return 13
	default:
		return log2(v)
	}
}

func log2[T Log2Types](v T) T {
	var r T
	for ; v > 1; v >>= 1 {
		r++
	}

	return r
}
