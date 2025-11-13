package tempvar

import (
	"encoding/binary"
	"errors"
	"math/rand/v2"
	"time"
)

var (
	ErrMaxMustNotBeZero  = errors.New("max must not be zero")
	ErrThresholdOverflow = errors.New("threshold overflow")
)

type Chance[T any] struct {
	rng    *rand.Rand
	max    uint
	thresh uint
	value  *T
}

func NewChance[T any](value T, max uint, thresh uint) (*Chance[T], error) {
	if max == 0 {
		return nil, ErrMaxMustNotBeZero
	}

	if thresh >= max {
		return nil, ErrThresholdOverflow
	}

	return &Chance[T]{
		rng:    setupRNG(),
		max:    max,
		thresh: thresh,
		value:  &value,
	}, nil
}

func (c *Chance[T]) Value() *T {
	if c.thresh > c.rng.UintN(c.max) {
		return nil
	}

	return c.value
}

func setupRNG() *rand.Rand {
	b := make([]byte, 32)

	now := uint64(time.Now().UnixMilli())

	binary.NativeEndian.PutUint64(b, now)
	binary.NativeEndian.PutUint64(b, now)
	binary.NativeEndian.PutUint64(b, now)
	binary.NativeEndian.PutUint64(b, now)

	return rand.New(rand.NewChaCha8([32]byte(b)))
}
