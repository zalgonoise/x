package memmap

import (
	"context"
	"errors"

	"github.com/zalgonoise/x/dns/store"
)

var (
	ErrDoesNotExist = errors.New("record does not exist")
)

type MemoryStore struct {
	// maps a set of domain names to record types to IPs
	Records map[string]map[string]string
}

func New() *MemoryStore {
	return &MemoryStore{
		Records: map[string]map[string]string{},
	}
}

func (m *MemoryStore) Add(ctx context.Context, rs ...store.Record) error {
	for _, r := range rs {
		dottedN := r.Name + "."

		if _, ok := m.Records[dottedN]; !ok {
			m.Records[dottedN] = map[string]string{}
		}
		m.Records[dottedN][r.Type] = r.Addr
	}
	return nil
}

func (m *MemoryStore) List(ctx context.Context) ([]store.Record, error) {
	var output []store.Record

	for domain, r := range m.Records {
		for rtype, addr := range r {
			output = append(output, store.Record{
				Type: rtype,
				Addr: addr,
				Name: domain,
			})
		}
	}
	return output, nil
}

func (m *MemoryStore) GetByAddr(ctx context.Context, rtype string, addr string) (store.Record, error) {
	if _, ok := m.Records[addr]; !ok {
		return store.Record{}, ErrDoesNotExist
	}

	dest := m.Records[addr][rtype]
	if dest == "" {
		return store.Record{}, ErrDoesNotExist
	}
	return store.Record{
		Type: rtype,
		Addr: dest,
		Name: addr,
	}, nil
}

func (m *MemoryStore) GetByDest(ctx context.Context, addr string) ([]store.Record, error) {
	var output []store.Record

	for domain, r := range m.Records {
		for rtype, ipAddr := range r {
			if addr == ipAddr {
				output = append(output, store.Record{
					Type: rtype,
					Addr: ipAddr,
					Name: domain,
				})
			}
		}
	}
	return output, nil
}

func (m *MemoryStore) Update(ctx context.Context, addr string, r store.Record) error {
	m.Records[addr+"."][r.Type] = r.Addr
	return nil
}

func (m *MemoryStore) Delete(ctx context.Context, addr string) error {
	for domain, r := range m.Records {
		if domain == addr+"." {
			for key := range r {
				r[key] = ""
			}
		}
	}
	return nil
}
