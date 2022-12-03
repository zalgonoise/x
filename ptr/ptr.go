package ptr

// From takes in a pointer to a type, and returns
// the value stored in the pointer, and a boolean on weather
// it is nil or not
//
// If the value is nil, it returns an initialized version of the
// type, regardless what value it is
func From[T any](ptr *T) (T, bool) {
	if ptr == nil {
		return *new(T), false
	}
	return *ptr, true
}

// Must will return the value stored in the input pointer to a type,
// if it is not nil
//
// If the value is nil, it returns an initialized version of the
// type, regardless what value it is
func Must[T any](ptr *T) T {
	if ptr == nil {
		return *new(T)
	}
	return *ptr
}

// To will return a pointer to the input value
func To[T any](value T) *T {
	return &value
}

// Copy will return a literal copy of the value stored in the input pointer
// as a pointer, too
func Copy[T any](ptr *T) *T {
	newT := new(T)
	*newT = *ptr
	return newT
}
