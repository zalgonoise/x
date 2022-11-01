package store

import "context"

// Repository defines the set of operations that a record store should expose
//
// This will consist of basic CRUD operations against a key-value store,
// to add, list, get, update or delete DNS Records from the key-value store.
//
// Additionally, it is exposing both GetByAddr and GetByDest methods to
// fetch items in the records list interchangeably
type Repository interface {
	// Create will add a new entry in they key-value store to include a
	// new Record, returning an error
	Create(context.Context, ...*Record) error

	// List will fetch all records in the key-value store
	List(context.Context) ([]*Record, error)

	// FilterByTypeAndDomain will fetch an address based on its address and type strings
	//
	// FilterByTypeAndDomain(ctx, "A", "service.mydomain") -> { "127.0.0.1", nil }
	FilterByTypeAndDomain(context.Context, string, string) (*Record, error)

	// FilterByDomain will fetch an address based on its address for all types
	//
	// FilterByDomain(ctx, "service.mydomain") -> { ["127.0.0.1"], nil }
	FilterByDomain(context.Context, string) ([]*Record, error)

	// FilterByDest will fetch an address based on its target IP
	//
	// FilterByDest(ctx, "127.0.0.1") -> { ["service.mydomain"], nil }
	FilterByDest(context.Context, string) ([]*Record, error)

	// Update will modify an existing record by targetting its domain string,
	// and by supplying a new version of the Record to update. Returns an error
	Update(context.Context, string, *Record) error

	// Delete will remove a DNS Record, filtering as per the provided data in the Record
	Delete(context.Context, *Record) error
}
