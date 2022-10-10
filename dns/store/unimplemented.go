package store

import (
	"context"
	"errors"
)

var (
	ErrUnimplemented error = errors.New("unimplemented DNS Record Store")
)

type unimplemented struct{}

func (u unimplemented) Add(context.Context, ...*Record) error {
	return ErrUnimplemented
}

func (u unimplemented) List(context.Context) ([]*Record, error) {
	return nil, ErrUnimplemented
}

func (u unimplemented) GetByDomain(context.Context, *Record) (*Record, error) {
	return nil, ErrUnimplemented
}

func (u unimplemented) GetByDest(context.Context, *Record) ([]*Record, error) {
	return nil, ErrUnimplemented
}

func (u unimplemented) Update(context.Context, string, *Record) error {
	return ErrUnimplemented
}

func (u unimplemented) Delete(context.Context, *Record) error {
	return ErrUnimplemented
}

func Unimplemented() unimplemented {
	return unimplemented{}
}
