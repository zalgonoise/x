package attr

func IntAttr[T IntRestriction](key string, value T) Attr {
	return NewAttr(key, (int64)(value))
}
func UintAttr[T UintRestriction](key string, value T) Attr {
	return NewAttr(key, (uint64)(value))
}
func FloatAttr[T FloatRestriction](key string, value T) Attr {
	return NewAttr(key, (float64)(value))
}
func ComplexAttr[T ComplexRestriction](key string, value T) Attr {
	return NewAttr(key, (complex128)(value))
}
func StringAttr[T CharRestriction](key string, value T) Attr {
	return NewAttr(key, (string)(value))
}
