package memmap

import (
	"context"

	"github.com/zalgonoise/x/dns/store"
)

// Create implements the store.Repository interface
//
// It will not perform any lookups before writing the new records, and it will simply
// blindly write the input records to the store.
func (m *MemoryStore) Create(ctx context.Context, rs ...*store.Record) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	for _, r := range rs {
		if _, ok := m.Records[r.Type]; !ok {
			m.Records[r.Type] = map[string]string{}
		}
		m.Records[r.Type][r.Name] = r.Addr
	}
	return nil
}

// List implements the store.Repository interface
//
// It will build a list of pointers to store.Record which is returned alongside
// any errors that are raised (currently there are no scenarios in this implementation)
func (m *MemoryStore) List(ctx context.Context) ([]*store.Record, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	var output []*store.Record

	for rtype, r := range m.Records {
		for domain, addr := range r {
			output = append(
				output,
				store.New().
					Type(rtype).
					Name(domain).
					Addr(addr).
					Build(),
			)
		}
	}
	return output, nil
}

// FilterByDomain implements the store.Repository interface
//
// It will return a pointer to a store.Record if there is an IP address
// registered to the input store.Record's domain name and record type.
//
// It also returns an error in case the record does not exist
func (m *MemoryStore) FilterByDomain(ctx context.Context, r *store.Record) (*store.Record, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if _, ok := m.Records[r.Type]; !ok {
		return nil, store.ErrDoesNotExist
	}
	dest := m.Records[r.Type][r.Name]
	if dest == "" {
		return nil, store.ErrDoesNotExist
	}

	r.Addr = dest
	return r, nil
}

// FilterByDest implements the store.Repository interface
//
// It will return a slice of pointers to store.Records if there are records
// associated with the input store.Record's IP address.
//
// If the call is successful but there are no records associated to that addres,
// returns an empty slice.
//
// It also returns an error in case the operation fails (which is currently not
// a scenario)
func (m *MemoryStore) FilterByDest(ctx context.Context, r *store.Record) ([]*store.Record, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	var output []*store.Record

	for rtype, domains := range m.Records {
		for domain, ipAddr := range domains {
			if ipAddr == r.Addr {
				output = append(
					output,
					store.New().
						Type(rtype).
						Name(domain).
						Addr(r.Addr).
						Build(),
				)
			}
		}
	}
	return output, nil
}

// Update implements the store.Repository interface
//
// It will target a particular domain name, and update its target IP address
// based on the input store.Record.
//
// If it targets a domain which does not exist in the store, or if that domain
// does not have that record type registered, it returns a DoesNotExist error
//
// If the operation is successful, it returns nil
func (m *MemoryStore) Update(ctx context.Context, domain string, r *store.Record) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if _, ok := m.Records[r.Type]; !ok {
		return store.ErrDoesNotExist
	}
	if _, ok := m.Records[r.Type][domain]; !ok {
		return store.ErrDoesNotExist
	}
	m.Records[r.Type][domain] = r.Addr
	return nil
}

// Delete implements the store.Repository interface
//
// It will leverage the input store.Record to commit to a record deletion.
//
// If:
//   - a domain name is provided: its target IP address and record types are removed
//   - a domain name and record type are provided: remove the target IP address associated
//   - IP address is populated: delete all records associated with that address
//
// It returns an error if the operation is unsuccessful (which is not a scenario as of now)
func (m *MemoryStore) Delete(ctx context.Context, r *store.Record) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if r.Name != "" && r.Type == "" && r.Addr == "" {
		deleteDomain(m, r.Name)
	}
	if r.Name != "" && r.Type != "" && r.Addr == "" {
		deleteDomainByType(m, r.Name, r.Type)
	}
	if r.Addr != "" {
		deleteAddress(m, r.Addr)
	}
	return nil
}
