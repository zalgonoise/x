package data

// BitDepthTypes is a type constraint joining all the different
// data types used by the supported bit depths
type BitDepthTypes interface {
	int8 | int16 | int32 | byte | int
}

func conv[F, T BitDepthTypes](a []F, steps int, fn func([]F) T) []T {
	out := make([]T, len(a)/steps)
	for i, j := 0, 0; i+steps-1 < len(a); i, j = i+steps, j+1 {
		out[j] = fn(a[i : i+steps])
	}
	return out
}

func to[F, T BitDepthTypes](from []F) []T {
	out := make([]T, len(from))
	for i := range from {
		out[i] = (T)(from[i])
	}
	return out
}

func encode24BitLE(buf []byte, v int32) []byte {
	return append(buf, byte(v), byte(v>>8), byte(v>>16))
}

func decode24BitLE(buf []byte) int32 {
	value := int32(buf[0]) | (int32(buf[1]) << 8) | (int32(buf[2]) << 16)
	if value&0x00800000 != 0 {
		value |= -16777216 // handle signed integers
	}
	return value
}

// can't inline a pointer cast and convert an array to a slice
func append2Bytes(idx int, dst []byte, src [2]byte) {
	if idx*2 < len(dst) {
		copy(dst[idx*2:], src[:])
	}
}

// can't inline a pointer cast and convert an array to a slice
func append3Bytes(idx int, dst []byte, src [3]byte) {
	copy(dst[idx*3:], src[:])
}

// can't inline a pointer cast and convert an array to a slice
func append4Bytes(idx int, dst []byte, src [4]byte) {
	copy(dst[idx*4:], src[:])
}
