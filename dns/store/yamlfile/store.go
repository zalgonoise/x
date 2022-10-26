package yamlfile

import (
	"context"
	"fmt"

	"github.com/zalgonoise/x/dns/store"
)

// Create implements the store.Repository interface
//
// It will call the in-memory store's method of the same signature, while deferring
// a `Sync()` call to ensure the records file is up-to-date
func (f *FileStore) Create(ctx context.Context, rs ...*store.Record) error {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	err := f.store.Create(ctx, rs...)
	if err != nil {
		return fmt.Errorf("failed to create record: %w", err)
	}
	err = f.sync()
	if err != nil {
		return fmt.Errorf("failed to sync store to file: %w", err)
	}
	return nil
}

// List implements the store.Repository interface
//
// It will call the in-memory store's method of the same signature
func (f *FileStore) List(ctx context.Context) ([]*store.Record, error) {
	rs, err := f.store.List(ctx)
	if err != nil {
		return rs, fmt.Errorf("failed to list records: %w", err)
	}
	return rs, nil
}

// FilterByDomain implements the store.Repository interface
//
// It will call the in-memory store's method of the same signature
func (f *FileStore) FilterByTypeAndDomain(ctx context.Context, rtype, domain string) (*store.Record, error) {
	r, err := f.store.FilterByTypeAndDomain(ctx, rtype, domain)
	if err != nil {
		return r, fmt.Errorf("failed to get record by domain: %w", err)
	}
	return r, nil
}

// FilterByDest implements the store.Repository interface
//
// It will call the in-memory store's method of the same signature
func (f *FileStore) FilterByDest(ctx context.Context, addr string) ([]*store.Record, error) {
	rs, err := f.store.FilterByDest(ctx, addr)
	if err != nil {
		return rs, fmt.Errorf("failed to get record by address: %w", err)
	}
	return rs, nil
}

// Update implements the store.Repository interface
//
// It will call the in-memory store's method of the same signature, while deferring
// a `Sync()` call to ensure the records file is up-to-date
func (f *FileStore) Update(ctx context.Context, domain string, r *store.Record) error {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	err := f.store.Update(ctx, domain, r)
	if err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}
	err = f.sync()
	if err != nil {
		return fmt.Errorf("failed to sync store to file: %w", err)
	}
	return nil
}

// Delete implements the store.Repository interface
//
// It will call the in-memory store's method of the same signature, while deferring
// a `Sync()` call to ensure the records file is up-to-date
func (f *FileStore) Delete(ctx context.Context, r *store.Record) error {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	err := f.store.Delete(ctx, r)
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	err = f.sync()
	if err != nil {
		return fmt.Errorf("failed to sync store to file: %w", err)
	}
	return nil
}
