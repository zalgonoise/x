package tempvar

import (
	"errors"
	"math/rand/v2"
	"unsafe"
)

var (
	ErrUnsupportedIntSize = errors.New("unsupported int size")
)

const (
	maxAddr32 = 0xffffffff
	maxAddr64 = 0xffffffffffffffff
)

type ImpostorAddr[T any] struct {
	rng     *rand.Rand
	max     uint
	thresh  uint
	value   *T
	maxAddr uint
}

func NewImpostorAddr[T any](value T, max, thresh uint, intSize int) (*ImpostorAddr[T], error) {
	if max == 0 {
		return nil, ErrMaxMustNotBeZero
	}

	if thresh >= max {
		return nil, ErrThresholdOverflow
	}

	var maxAddr uint
	switch intSize {
	case 32:
		maxAddr = maxAddr32
	case 64:
		maxAddr = maxAddr64
	default:
		return nil, ErrUnsupportedIntSize
	}

	return &ImpostorAddr[T]{
		rng:     setupRNG(),
		max:     max,
		thresh:  thresh,
		value:   &value,
		maxAddr: maxAddr,
	}, nil
}

func (c *ImpostorAddr[T]) Value() uintptr {
	if c.thresh > c.rng.UintN(c.max) {
		return uintptr(c.rng.UintN(c.maxAddr))
	}

	return uintptr(unsafe.Pointer(c.value))
}
