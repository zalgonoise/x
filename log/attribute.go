package log

import (
	"errors"
	"time"
)

type Attr interface {
	Key() string
	Value() any
	WithKey(key string) Attr
	WithValue(value any) Attr
}

func NewAttr[T any](key string, value T) Attr {
	if err, ok := (any)(value).(error); ok {
		return ErrAttr(key, err)
	}
	return attr[T]{
		key:   key,
		value: value,
	}
}

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
func ErrAttr(key string, err error) Attr {
	var errs []string
	for err != nil {
		errs = append(errs, err.Error())
		err = errors.Unwrap(err)
	}
	if len(errs) == 1 {
		return NewAttr(key, errs[0])
	}
	return NewAttr(key, errs)
}

type attr[T any] struct {
	key   string
	value T
}

func (a attr[T]) Key() string {
	return a.key
}

func (a attr[T]) Value() any {
	return a.value
}

func (a attr[T]) WithKey(key string) Attr {
	return NewAttr(key, a.value)
}

func (a attr[T]) WithValue(value any) Attr {
	if v, ok := (value).(T); ok {
		return NewAttr(a.key, v)
	}
	return nil
}

type IntRestriction interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}
type UintRestriction interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}
type FloatRestriction interface {
	~float32 | ~float64
}
type ComplexRestriction interface {
	~complex64 | ~complex128
}
type NumberRestriction interface {
	IntRestriction | UintRestriction | FloatRestriction | ComplexRestriction
}
type CharRestriction interface {
	~string | ~byte | ~rune | ~[]byte | ~[]rune
}
type BoolRestriction interface {
	~bool
}
type TimeRestriction interface {
	time.Time
}
type TextRestriction interface {
	CharRestriction | BoolRestriction | TimeRestriction
}
