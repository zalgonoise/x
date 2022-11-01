package store

import (
	"context"
	"errors"
)

var (
	ErrUnimplemented error = errors.New("unimplemented DNS Record Store")
)

type unimplemented struct{}

// Create implements the store.Repository interface
func (u unimplemented) Create(ctx context.Context, rs ...*Record) error {
	return ErrUnimplemented
}

// List implements the store.Repository interface
func (u unimplemented) List(ctx context.Context) ([]*Record, error) {
	return nil, ErrUnimplemented
}

// FilterByTypeAndDomain implements the store.Repository interface
func (u unimplemented) FilterByTypeAndDomain(ctx context.Context, rtype, domain string) (*Record, error) {
	return nil, ErrUnimplemented
}

// FilterByDomain implements the store.Repository interface
func (u unimplemented) FilterByDomain(ctx context.Context, domain string) ([]*Record, error) {
	return nil, ErrUnimplemented
}

// FilterByDest implements the store.Repository interface
func (u unimplemented) FilterByDest(ctx context.Context, addr string) ([]*Record, error) {
	return nil, ErrUnimplemented
}

// Update implements the store.Repository interface
func (u unimplemented) Update(ctx context.Context, domain string, r *Record) error {
	return ErrUnimplemented
}

// Delete implements the store.Repository interface
func (u unimplemented) Delete(ctx context.Context, r *Record) error {
	return ErrUnimplemented
}

// Unimplemented returns an unimplemented (and invalid) store.Repository
func Unimplemented() unimplemented {
	return unimplemented{}
}
