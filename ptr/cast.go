package ptr

import (
	"unsafe"
)

// Cast converts the input type From in `value` into To type, using an unsafe approach
//
// While this approach is amazingly fast, it's very strict in the sense that the From and To
// types must match exactly.
//
// If From and To are custom types like structs, the element types and field order must be exactly the same,
// otherwise the runtime will panic. There is no safe way to check for this except to "try it first".
func Cast[To, From any](value From) To {
	return *(*To)(unsafe.Pointer(&value))
}

// CastPtr is just like Cast, but accepts a pointer to a From type in `value`, and also
// returns a pointer to To.
func CastPtr[To, From any](value *From) *To {
	return (*To)(unsafe.Pointer(value))
}
