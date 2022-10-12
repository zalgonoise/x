package store

import "context"

// Repository defines the behavior that a record store should have
//
// This will consist of basic CRUD operations against a key-value store,
// to add, list, get, update or delete DNS Records from the key-value store.
//
// Additionally, it is exposing both GetByAddr and GetByDest methods to
// fetch items in the records list interchangeably
type Repository interface {
	// Add will create a new entry in they key-value store to include a
	// new Record, returning an error
	Add(context.Context, ...*Record) error

	// List will fetch all records in the key-value store
	//
	// TODO: implement listing filters
	List(context.Context) ([]*Record, error)

	// GetByDomain will fetch an address based on its address string
	//
	// GetByDomain(ctx, "service.mydomain") -> { "127.0.0.1", nil }
	GetByDomain(context.Context, *Record) (*Record, error)

	// GetByDest will fetch an address based on its target IP
	//
	// GetByDest(ctx, "127.0.0.1") -> { ["service.mydomain"], nil }
	GetByDest(context.Context, *Record) ([]*Record, error)

	// Update will modify an existing record by targetting its domain string,
	// and by supplying a new version of the Record to update. Returns an error
	Update(context.Context, string, *Record) error

	// Delete will remove a DNS Record, filtering as per the provided data in the Record
	Delete(context.Context, *Record) error
}
