package attr

func Int[T IntRestriction](key string, value T) Attr {
	return New(key, (int64)(value))
}
func Uint[T UintRestriction](key string, value T) Attr {
	return New(key, (uint64)(value))
}
func Float[T FloatRestriction](key string, value T) Attr {
	return New(key, (float64)(value))
}
func Complex[T ComplexRestriction](key string, value T) Attr {
	return New(key, (complex128)(value))
}
func String[T CharRestriction](key string, value T) Attr {
	return New(key, (string)(value))
}
