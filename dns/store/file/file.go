package file

import (
	"context"
	"io"
	"log"
	"os"
	"sync"

	"github.com/zalgonoise/x/dns/store"
	"github.com/zalgonoise/x/dns/store/memmap"
)

// FileStore is an in-memory implementation of a DNS record store
// wrapped with a syncer that will dump / retrieve DNS record data
// from a file in JSON format
//
// The in-memory implementation used is store/memmap
type FileStore struct {
	Path  string `json:"path,omitempty" yaml:"path,omitempty"`
	store store.Repository
	enc   EncodeDecoder
	mtx   sync.RWMutex
}

// Store holds a set of (DNS) Records
type Store struct {
	Types []*Type `json:"types,omitempty" yaml:"types,omitempty"`
}

// Record is labeled by an IP address and contains a slice of (pointers to) Types
type Record struct {
	Address string   `json:"address,omitempty" yaml:"address,omitempty"`
	Domains []string `json:"domains,omitempty"   yaml:"domains,omitempty"`
}

// Type is labeled by a DNS record type and contains a slice of Domains
type Type struct {
	RType   string    `json:"type,omitempty"    yaml:"type,omitempty"`
	Records []*Record `json:"records,omitempty" yaml:"records,omitempty"`
}

// New returns a new JSON FileStore as a store.Repository
//
// It takes in a path to a file which will be used for reads and writes,
// to back-up and sync the record store to disk.
//
// This initialization function will try to open an existing file, or create it
// if it does not exist, and also read it if it has content. If any of the critical
// operations fail, the function will panic since the store will not be able to start.
//
// TODO: decide if it's better to return a naked in-memory record store and log as critical
func New(encoderType, path string) store.Repository {
	var (
		mainEnc     EncodeDecoder
		altEnc      EncodeDecoder
		mainEncType string
		altEncType  string
	)
	switch encoderType {
	case "json":
		mainEncType = "json"
		altEncType = "yaml"
	case "yaml":
		mainEncType = "yaml"
		altEncType = "json"
	default:
		mainEncType = "yaml"
		altEncType = "json"
	}

	mainEnc = NewEncoder(mainEncType)
	altEnc = NewEncoder(altEncType)

	mstore := memmap.New()
	f, err := os.OpenFile(path, os.O_CREATE, os.FileMode(store.OS_ALL_RW))
	if err != nil {
		panic(err) // panic on init if file can't be saved to disk
	}
	b, err := io.ReadAll(f)
	if err != nil {
		panic(err) // panic on init if file can't be opened / used
	}
	if len(b) > 0 {
		s := &Store{}
		meErr := mainEnc.Decode(b, s)
		if meErr != nil {
			aeErr := altEnc.Decode(b, s)
			if aeErr != nil {
				log.Printf("failed to unmarshal %s: %v ; failed to unmarshal %s: %v\n", mainEncType, meErr, altEncType, aeErr)
			}
		}

		err := mstore.Create(context.Background(), toEntity(s)...)
		if err != nil {
			log.Printf("error adding entries: %v\n", err)
		}
	}

	return &FileStore{
		Path:  path,
		store: mstore,
		enc:   mainEnc,
	}
}
