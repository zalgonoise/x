package store

import (
	"context"
)

// Record defines the elements of a DNS Record
//
// TODO: add the elements necessary to comprehend the most common
// DNS records' elements
type Record struct {
	Type string
	Name string
	Addr string
}

// StoreRepository defines the behavior that a record store should have
//
// This will consist of basic CRUD operations against a key-value store,
// to add, list, get, update or delete DNS Records from the key-value store.
//
// Additionally, it is exposing both GetByAddr and GetByDest methods to
// fetch items in the records list interchangeably
type Repository interface {
	// Add will create a new entry in they key-value store to include a
	// new Record, returning an error
	Add(ctx context.Context, r ...Record) error

	// List will fetch all records in the key-value store
	//
	// TODO: implement listing filters
	List(ctx context.Context) ([]Record, error)

	// GetByAddr will fetch an address based on its address string
	//
	// GetByAddr(ctx, "service.mydomain") -> { "127.0.0.1", nil }
	GetByAddr(ctx context.Context, rtype string, addr string) (Record, error)

	// GetByDest will fetch an address based on its target IP
	//
	// GetByDest(ctx, "127.0.0.1") -> { ["service.mydomain"], nil }
	GetByDest(ctx context.Context, addr string) ([]Record, error)

	// Update will modify an existing record by targetting its address string,
	// and by supplying a new version of the Record to update. Returns an error
	Update(ctx context.Context, addr string, r Record) error

	// Delete will remove a DNS Record by targetting its address string
	Delete(ctx context.Context, addr string) error
}
