package memmap

import (
	"sync"

	"github.com/zalgonoise/x/dns/store"
)

// MemoryStore is an in-memory implementation of a DNS record store
//
// It uses simple Go maps to represent a relationship of
// record-type-to-domain-to-IP as a map[string]map[string]string.
//
// This direction is so that DNS queries can be answered faster, while the remaining
// operations are not as important.
//
// It also has a sync.RWMutex to ensure that data races do not occur
type MemoryStore struct {
	// maps a set of domain names to record types to IPs
	Records map[string]map[string]string
	mtx     sync.RWMutex
}

// New returns a new MemoryStore as a store.Repository
func New() store.Repository {
	return &MemoryStore{
		Records: map[string]map[string]string{},
	}
}
