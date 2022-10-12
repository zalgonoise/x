package memmap

import (
	"context"
	"errors"
	"sync"

	"github.com/zalgonoise/x/dns/store"
)

var (
	ErrDoesNotExist = errors.New("record does not exist")
	ErrNoAddr       = errors.New("no IP address provided")
	ErrNoName       = errors.New("no domain name provided")
	ErrNoType       = errors.New("no DNS record type provided")
)

type MemoryStore struct {
	// maps a set of domain names to record types to IPs
	Records map[string]map[string]string
	mtx     sync.RWMutex
}

func New() *MemoryStore {
	return &MemoryStore{
		Records: map[string]map[string]string{},
	}
}

func (m *MemoryStore) Add(ctx context.Context, rs ...*store.Record) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	for _, r := range rs {
		dottedN := r.Name + "."

		if _, ok := m.Records[dottedN]; !ok {
			m.Records[dottedN] = map[string]string{}
		}
		m.Records[dottedN][r.Type] = r.Addr
	}
	return nil
}

func (m *MemoryStore) List(ctx context.Context) ([]*store.Record, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	var output []*store.Record

	for domain, r := range m.Records {
		for rtype, addr := range r {
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

func (m *MemoryStore) GetByDomain(ctx context.Context, r *store.Record) (*store.Record, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if _, ok := m.Records[r.Name]; !ok {
		return nil, ErrDoesNotExist
	}
	dest := m.Records[r.Name][r.Type]
	if dest == "" {
		return nil, ErrDoesNotExist
	}

	r.Addr = dest
	return r, nil
}

func (m *MemoryStore) GetByDest(ctx context.Context, r *store.Record) ([]*store.Record, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	var output []*store.Record

	for domain, rmap := range m.Records {
		for rtype, ipAddr := range rmap {
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

func (m *MemoryStore) Update(ctx context.Context, domain string, r *store.Record) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	m.Records[domain+"."][r.Type] = r.Addr
	return nil
}

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
