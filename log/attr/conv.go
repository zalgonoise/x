package attr

// Int is a generic function to create Attr attributes based on
// int values. It converts the values to int64
func Int[T IntRestriction](key string, value T) Attr {
	return New(key, (int64)(value))
}

// Uint is a generic function to create Attr attributes based on
// uint values. It converts the values to uint64
func Uint[T UintRestriction](key string, value T) Attr {
	return New(key, (uint64)(value))
}

// Float is a generic function to create Attr attributes based on
// float values. It converts the values to float64
func Float[T FloatRestriction](key string, value T) Attr {
	return New(key, (float64)(value))
}

// Complex is a generic function to create Attr attributes based on
// complex values. It converts the values to complex128
func Complex[T ComplexRestriction](key string, value T) Attr {
	return New(key, (complex128)(value))
}

// String is a generic function to create Attr attributes based on
// string values. It converts the values to string
func String[T CharRestriction](key string, value T) Attr {
	return New(key, (string)(value))
}
