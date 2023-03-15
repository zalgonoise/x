package ptr

import (
	"errors"
	"unsafe"
)

var ErrInvalidCap = errors.New("invalid slice capacity")

const (
	Cap1 int = 1 << iota
	Cap2
	Cap4
	Cap8
	Cap16
	Cap32
	Cap64
	Cap128
	Cap256
	Cap512
	Cap1024
	Cap2048
	Cap4096
	Cap8192
)

// ToArray converts a slice into an array of the corresponding size,
// with a pointer manipulation method (no copying, no reflection)
//
// The caveats to using this function is the limited scope it provides.
// Since arrays in Go can only be defined with a constant value, one cannot
// simply generate "an array of any capacity" from a function call -- only from
// a capacity defined in a constant. Example:
//
//	// this will not work
//	func[T any, C int](slice []T) ([C]T, error)
//
//	// this will not work
//	var size = len(slice)
//	var arr = [size]T{}
//
//	// this will work (implying you know the capacity)
//	const size = 3
//	var arr = [size]T{}
//
// For the same reason, the function is scoped to a set of log(2) constants:
// 1, 2, 4, 8, 16, 32, 64, 128, 156, 512, 1024, 2048, 4096, 8192.
//
// If the user wants a custom value, it's best just to copy the two actions that really matter:
//
// 1. Getting the slice metadata as a [3]int value; that hold the pointer to the array, the slice length
// and slice capacity: `sliceData := *(*[3]uint)(unsafe.Pointer(&slice))`
//
// 2. Ensuring the constant is the same value as the slice capacity (return an error if it doesn't match)
//
// 3. Return the first value of the slice, cast as an array of the needed size:
// *(*[{your-capacity}]T)(unsafe.Pointer(uintptr(sliceData[0])))
//
// So, if you needed to convert a 4-item-long ints slice into a 4-item-array, you could do so, and expect:
//
//	ints := []int{1, 2, 3, 4}
//	arr, _ := ptr.ToArray(ints) // skip error checks for example
//
// You would expect the arr variable to be `[4]int{1, 2, 3, 4}`
func ToArray[T any](slice []T) (any, error) {
	// sliceData holds:
	// [0]: pointer to the underlying array
	// [1]: slice length
	// [2]: slice capacity
	sliceData := *(*[3]uint)(unsafe.Pointer(&slice))

	// the pointer in sliceData[0] can be casted to an array only if the
	// identifier for the array capacity is a constant.
	//
	// due to this, it's only possible to convert a slice to an array with
	// unsafe.Pointer if the array size is predeclared in this package. For
	// convenience, it exposes constants for a log(2) set of values up to 8192
	// but this switch statement still needs to happen in order to use those
	// values as constants.
	//
	// the switch statement is looking into the slice's capacity and casting the
	// data pointer with the corresponding constant if it exists
	switch sliceData[2] {
	case 1:
		return *(*[Cap1]T)(unsafe.Pointer(uintptr(sliceData[0]))), nil
	case 2:
		return *(*[Cap2]T)(unsafe.Pointer(uintptr(sliceData[0]))), nil
	case 4:
		return *(*[Cap4]T)(unsafe.Pointer(uintptr(sliceData[0]))), nil
	case 8:
		return *(*[Cap8]T)(unsafe.Pointer(uintptr(sliceData[0]))), nil
	case 16:
		return *(*[Cap16]T)(unsafe.Pointer(uintptr(sliceData[0]))), nil
	case 32:
		return *(*[Cap32]T)(unsafe.Pointer(uintptr(sliceData[0]))), nil
	case 64:
		return *(*[Cap64]T)(unsafe.Pointer(uintptr(sliceData[0]))), nil
	case 128:
		return *(*[Cap128]T)(unsafe.Pointer(uintptr(sliceData[0]))), nil
	case 256:
		return *(*[Cap256]T)(unsafe.Pointer(uintptr(sliceData[0]))), nil
	case 512:
		return *(*[Cap512]T)(unsafe.Pointer(uintptr(sliceData[0]))), nil
	case 1024:
		return *(*[Cap1024]T)(unsafe.Pointer(uintptr(sliceData[0]))), nil
	case 2048:
		return *(*[Cap2048]T)(unsafe.Pointer(uintptr(sliceData[0]))), nil
	case 4096:
		return *(*[Cap4096]T)(unsafe.Pointer(uintptr(sliceData[0]))), nil
	case 8192:
		return *(*[Cap8192]T)(unsafe.Pointer(uintptr(sliceData[0]))), nil
	default:
		return nil, ErrInvalidCap
	}
}
