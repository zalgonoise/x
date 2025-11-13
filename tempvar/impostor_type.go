package tempvar

import (
	"math/rand/v2"
	"unsafe"
)

var (
	impostorString     string          = "kinda sus"
	impostorInt        int             = 420
	impostorUint       uint            = 420
	impostorInt8       int8            = 7
	impostorUint8      uint8           = 7
	impostorInt16      int16           = 420
	impostorUint16     uint16          = 420
	impostorInt32      int32           = 420
	impostorUint32     uint32          = 420
	impostorInt64      int64           = 420
	impostorUint64     uint64          = 420
	impostorFloat32    float32         = 4.20
	impostorFloat64    float64         = 4.20
	impostorBool       bool            = true
	impostorByte       byte            = 10
	impostorRune       rune            = 10
	impostorComplex    complex64       = complex64(4.20)
	impostorComplex128 complex128      = complex128(4.20)
	impostorAny        any             = any("kinda sus")
	impostorByteSlice  []byte          = []byte("kinda sus")
	impostorByteArray  [9]byte         = [9]byte([]byte("kinda sus"))
	impostorMap        map[string]bool = map[string]bool{
		"kinda sus": true,
	}
	impostorStruct impostor      = impostor{"kinda sus"}
	impostorFunc   func() string = func() string {
		return "kinda sus"
	}
	impostorInterface     interface{ Value() string } = impostor{"kinda sus"}
	impostorPointer       *impostor                   = &impostor{"kinda sus"}
	impostorUnsafePointer unsafe.Pointer              = unsafe.Pointer(&impostor{tag: "kinda sus"})
	impostorUintptr       uintptr                     = uintptr(unsafe.Pointer(&impostor{tag: "kinda sus"}))
)

var (
	lenImpostors = uint(len(impostors))
	impostors    = [...]unsafe.Pointer{
		unsafe.Pointer(&impostorString),
		unsafe.Pointer(&impostorInt),
		unsafe.Pointer(&impostorUint),
		unsafe.Pointer(&impostorInt8),
		unsafe.Pointer(&impostorUint8),
		unsafe.Pointer(&impostorInt16),
		unsafe.Pointer(&impostorUint16),
		unsafe.Pointer(&impostorInt32),
		unsafe.Pointer(&impostorUint32),
		unsafe.Pointer(&impostorInt64),
		unsafe.Pointer(&impostorUint64),
		unsafe.Pointer(&impostorFloat32),
		unsafe.Pointer(&impostorFloat64),
		unsafe.Pointer(&impostorBool),
		unsafe.Pointer(&impostorByte),
		unsafe.Pointer(&impostorRune),
		unsafe.Pointer(&impostorComplex),
		unsafe.Pointer(&impostorComplex128),
		unsafe.Pointer(&impostorAny),
		unsafe.Pointer(&impostorByteSlice),
		unsafe.Pointer(&impostorByteArray),
		unsafe.Pointer(&impostorMap),
		unsafe.Pointer(&impostorStruct),
		unsafe.Pointer(&impostorFunc),
		unsafe.Pointer(&impostorInterface),
		unsafe.Pointer(&impostorPointer),
		unsafe.Pointer(&impostorUnsafePointer),
		unsafe.Pointer(&impostorUintptr),
	}
)

type ImpostorType[T any] struct {
	rng    *rand.Rand
	max    uint
	thresh uint
	value  *T
}

func NewImpostorType[T any](value T, max, thresh uint) (*ImpostorType[T], error) {
	if max == 0 {
		return nil, ErrMaxMustNotBeZero
	}

	if thresh >= max {
		return nil, ErrThresholdOverflow
	}

	return &ImpostorType[T]{
		rng:    setupRNG(),
		max:    max,
		thresh: thresh,
		value:  &value,
	}, nil
}

func (c *ImpostorType[T]) Value() unsafe.Pointer {
	if c.thresh > c.rng.UintN(c.max) {
		idx := c.rng.UintN(lenImpostors)

		return impostors[idx]
	}

	return unsafe.Pointer(c.value)
}

type impostor struct {
	tag string
}

func (i impostor) Value() string {
	return i.tag
}
