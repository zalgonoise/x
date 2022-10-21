package jsonfile

import (
	"context"

	"github.com/zalgonoise/x/dns/store"
)

// Create implements the store.Repository interface
//
// It will call the in-memory store's method of the same signature, while deferring
// a `Sync()` call to ensure the records file is up-to-date
func (f *FileStore) Create(ctx context.Context, rs ...*store.Record) error {
	f.mtx.Lock()
	defer func() {
		_ = f.sync()
	}()
	defer f.mtx.Unlock()

	return f.store.Create(ctx, rs...)
}

// List implements the store.Repository interface
//
// It will call the in-memory store's method of the same signature
func (f *FileStore) List(ctx context.Context) ([]*store.Record, error) {
	return f.store.List(ctx)
}

// FilterByDomain implements the store.Repository interface
//
// It will call the in-memory store's method of the same signature
func (f *FileStore) FilterByDomain(ctx context.Context, r *store.Record) (*store.Record, error) {
	return f.store.FilterByDomain(ctx, r)
}

// FilterByDest implements the store.Repository interface
//
// It will call the in-memory store's method of the same signature
func (f *FileStore) FilterByDest(ctx context.Context, r *store.Record) ([]*store.Record, error) {
	return f.store.FilterByDest(ctx, r)
}

// Update implements the store.Repository interface
//
// It will call the in-memory store's method of the same signature, while deferring
// a `Sync()` call to ensure the records file is up-to-date
func (f *FileStore) Update(ctx context.Context, domain string, r *store.Record) error {
	f.mtx.Lock()
	defer func() {
		_ = f.sync()
	}()
	defer f.mtx.Unlock()

	return f.store.Update(ctx, domain, r)
}

// Delete implements the store.Repository interface
//
// It will call the in-memory store's method of the same signature, while deferring
// a `Sync()` call to ensure the records file is up-to-date
func (f *FileStore) Delete(ctx context.Context, r *store.Record) error {
	f.mtx.Lock()
	defer func() {
		_ = f.sync()
	}()
	defer f.mtx.Unlock()

	return f.store.Delete(ctx, r)
}
