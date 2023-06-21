package ptr

import "unsafe"

// ABIType is a representation of Go ABI's (general / generic) type definition, as per internal/abi/type.go
//
// https://github.com/golang/go/blob/master/src/internal/abi/type.go#L20
type ABIType struct {
	Size_       uintptr
	PtrBytes    uintptr // number of (prefix) bytes in the type that can contain pointers
	Hash        uint32  // hash of type; avoids computation in hash tables
	TFlag       uint8   // extra type information flags
	Align_      uint8   // alignment of variable with this type
	FieldAlign_ uint8   // alignment of struct field with this type
	Kind_       uint8   // enumeration for C
	// function for comparing objects of this type
	// (ptr to object A, ptr to object B) -> ==?
	Equal func(unsafe.Pointer, unsafe.Pointer) bool
	// GCData stores the GC type data for the garbage collector.
	// If the KindGCProg bit is set in kind, GCData is a GC program.
	// Otherwise it is a ptrmask bitmap. See mbitmap.go for details.
	GCData    *byte
	Str       int32 // string form
	PtrToThis int32 // type for pointer to this type, may be zero
}

// ABIUncommonType is a representation of Go ABI's uncommon types, which are those that
// also contain methods, as per internal/abi/type.go
//
// ABIUncommonType is present only for defined types or types with methods
// (if T is a defined type, the uncommonTypes for T and *T have methods).
// Using a pointer to this struct reduces the overall size required
// to describe a non-defined type with no methods.
//
// https://github.com/golang/go/blob/master/src/internal/abi/type.go#L197
type ABIUncommonType struct {
	PkgPath int32  // import path; empty for built-in types like int, string
	Mcount  uint16 // number of methods
	Xcount  uint16 // number of exported methods
	Moff    uint32 // offset from this uncommontype to [mcount]Method
	_       uint32 // unused
}

// ABIName is a representation of Go ABI's name type, as per internal/abi/type.go
//
// https://github.com/golang/go/blob/master/src/internal/abi/type.go#L589
type ABIName struct {
	Bytes *byte
}

// ABIImethod is a representation of Go ABI's interface method type, as per internal/abi/type.go
//
// https://github.com/golang/go/blob/master/src/internal/abi/type.go#L231
type ABIImethod struct {
	Name int32 // name of method
	Typ  int32 // .(*FuncType) underneath
}

// ABIInterface is a representation of Go ABI's interface type, as per internal/abi/type.go
//
// https://github.com/golang/go/blob/master/src/internal/abi/type.go#L415
type ABIInterface struct {
	Type    ABIType
	PkgPath ABIName      // import path
	Methods []ABIImethod // sorted by hash
}

func (t *ABIType) Uncommon() *ABIUncommonType {
	type u struct {
		ABIInterface
		u ABIUncommonType
	}

	return &(*u)(unsafe.Pointer(t)).u
}

// Itable is a representation behind Go interface tables, as per runtime/runtime2.go
//
// https://github.com/golang/go/blob/master/src/runtime/runtime2.go#L951
type Itable struct {
	Inter ABIInterface
	Type  ABIType
	// Hash is a copy of Type.hash. Used for type switches.
	Hash uint32
	_    [4]byte
	// Fun is variable sized. fun[0]==0 means _type does not implement inter.
	Fun [1]uintptr
}

// Interface is a representation behind Go interfaces, as per runtime/runtime2.go
//
// https://github.com/golang/go/blob/master/src/runtime/runtime2.go#L204
type Interface struct {
	Itab  *Itable
	Value unsafe.Pointer
}

// GetInterface extracts the interface information from the input value. This input must be an interface, otherwise the
// results from this function call are unexpected, considering that it uses an unsafe approach to extract the underlying
// metadata and concrete type.
func GetInterface[T any](value T) Interface {
	return *(*Interface)(unsafe.Pointer(&value))
}

// IsNil returns a boolean if the supplied value is nil, inspected as an interface, using a
// fast (and unsafe) approach, by casting the input type as a Go interface, to check if its underlying
// concrete value is or isn't nil.
//
// This will also return true if the input value is an interface with a nil concrete type such as `fmt.Stringer(nil)`.
// It is up to the caller to evaluate if this is in fact useful within the logic of their application
//
// For this reason, the input type is `any`, an alias for `interface{}`. By casting the input as in interface in the
// function's parameters, we can use this pattern for many more data types besides interfaces.
//
// NOTE: this call will return invalid / incorrect results when it comes to slices. For that, please see the
// Slice definition, as its structure positions how the second element in the type cast operation would actually point
// to the slice's length; as such a nil slice would not present this second element as a zero-value uintptr (nil).
// It is up to the caller to avoid passing (nil) slices to this function to avoid these false negatives
func IsNil(value any) bool {
	if value == nil {
		return true
	}

	// fastest means of checking an interface's concrete type
	ptr := *(*[2]uintptr)(unsafe.Pointer(&value))

	return ptr[1] == 0
}

// IsEqual compares two interfaces via their built-in ABIType.Equal function
func IsEqual(ifaceA, ifaceB any) bool {
	if ifaceA == nil || ifaceB == nil {
		return false
	}

	return GetInterface(ifaceA).Itab.Type.Equal(
		(*[2]unsafe.Pointer)(unsafe.Pointer(&ifaceA))[1],
		(*[2]unsafe.Pointer)(unsafe.Pointer(&ifaceB))[1],
	)
}

// Match compares two interfaces for matching hashes, meaning that they should be
// compatible with one another (same list of methods)
func Match(ifaceA, ifaceB any) bool {
	if ifaceA == nil || ifaceB == nil {
		return false
	}

	i1 := GetInterface(ifaceA)
	i2 := GetInterface(ifaceB)

	return i1.Itab.Inter.Type.Hash == i2.Itab.Inter.Type.Hash
}
