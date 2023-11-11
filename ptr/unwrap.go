package ptr

import (
	"errors"
	"unsafe"
)

var (
	ErrNilInput = errors.New("nil input")
	ErrMaxDepth = errors.New("maximum depth reached")
)

// Unwrap traverses nested data structures to find a given type, that is wrapped as the first
// element of other (decorator) types.
//
// For instance, if a given service exposes a withLogs, withMetrics, and / or withTrace types,
// with a struct format that places the service as the first element and the other component as the second,
// this function will be able to extract that value in search for a given type T.
//
// Unwrap checks if the input value is of type T, and returns it if so. If it is not, then it uses
// an unsafe approach casting the input value as a struct containing an element of type T (as the first field).
// This retrieved element is then passed back into a private unwrap function that keeps looking for the type T
// until the maximum depth of 64 is reached.
//
// This technique allows to blindly extract a type from multiple nested decorators; if you're positive that the
// fields are there and ordered in the expected way.
func Unwrap[T any](value any) (T, error) {
	return unwrap[T](value, 0)
}

func unwrap[T any](value any, iter int) (T, error) {
	if iter >= maxDepth {
		return *new(T), ErrMaxDepth
	}

	switch v := value.(type) {
	case nil:
		return *new(T), ErrNilInput
	case T:
		return v, nil
	default:
		cast := (*struct{ x T })(unsafe.Pointer(&value))

		iter++
		return unwrap[T](cast.x, iter)
	}
}
