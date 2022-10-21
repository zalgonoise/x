package store

import (
	"context"
	"errors"
)

var (
	ErrUnimplemented error = errors.New("unimplemented DNS Record Store")
)

type unimplemented struct{}

// Create implements store.Repository
func (u unimplemented) Create(ctx context.Context, rs ...*Record) error {
	return ErrUnimplemented
}

// List implements store.Repository
func (u unimplemented) List(ctx context.Context) ([]*Record, error) {
	return nil, ErrUnimplemented
}

// FilterByDomain implements store.Repository
func (u unimplemented) FilterByDomain(ctx context.Context, r *Record) (*Record, error) {
	return nil, ErrUnimplemented
}

// FilterByDest implements store.Repository
func (u unimplemented) FilterByDest(ctx context.Context, r *Record) ([]*Record, error) {
	return nil, ErrUnimplemented
}

// Update implements store.Repository
func (u unimplemented) Update(ctx context.Context, domain string, r *Record) error {
	return ErrUnimplemented
}

// Delete implements store.Repository
func (u unimplemented) Delete(ctx context.Context, r *Record) error {
	return ErrUnimplemented
}

// Unimplemented returns an unimplemented (and invalid) store.Repository
func Unimplemented() unimplemented {
	return unimplemented{}
}
