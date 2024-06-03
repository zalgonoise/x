package lpc

type intSet interface {
	int | int8 | int16 | int32 | int64
}

type uintSet interface {
	uint | uint8 | uint16 | uint32 | uint64
}

func zigZag[To uintSet, From intSet](value From) To {
	switch {
	case value == 0:
		return 1
	case value < 0:
		return 2 * To(-value)
	default:
		return 2*To(value) + 1
	}
}
