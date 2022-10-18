package yamlfile

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"sync"

	"github.com/go-yaml/yaml"
	"github.com/zalgonoise/x/dns/store"
	"github.com/zalgonoise/x/dns/store/memmap"
)

var (
	ErrZeroBytesWritten error = errors.New("zero bytes written")
	ErrAlreadyExists    error = errors.New("entry already exists")
	ErrNotFound         error = errors.New("entry was not found")
	ErrSync             error = errors.New("sync error")
)

type FileStore struct {
	Path  string `yaml:"path,omitempty"`
	store store.Repository
	mtx   sync.RWMutex
}

type Store struct {
	Records []*Record `yaml:"records,omitempty"`
}

type Record struct {
	Address string  `yaml:"address,omitempty"`
	Types   []*Type `yaml:"types,omitempty"`
}

type Type struct {
	RType   string   `yaml:"type,omitempty"`
	Domains []string `yaml:"domains,omitempty"`
}

func New(path string) store.Repository {
	store := memmap.New()
	f, err := os.Open(path)
	if err != nil {
		panic(err) // panic on init if file can't be opened / used
	}
	b, err := io.ReadAll(f)
	if err != nil {
		panic(err) // panic on init if file can't be opened / used
	}
	if len(b) > 0 {
		s := &Store{}
		err = yaml.Unmarshal(b, s)
		if err != nil {
			log.Printf("failed to unmarshal YAML: %v\n", err)
		}

		err := store.Add(context.Background(), toEntity(s)...)
		if err != nil {
			log.Printf("error adding entries: %v\n", err)
		}
	}

	return &FileStore{
		Path:  path,
		store: store,
	}
}

func toEntity(s *Store) []*store.Record {
	var out []*store.Record

	for _, record := range s.Records {
		addr := record.Address
		for _, recordType := range record.Types {
			rtype := recordType.RType
			for _, domain := range recordType.Domains {
				out = append(out, store.New().
					Name(domain).
					Type(rtype).
					Addr(addr).
					Build())
			}
		}
	}

	return out
}

func fromEntity(rs ...*store.Record) *Store {
	out := &Store{}

inputLoop:
	for _, r := range rs {
		for _, record := range out.Records {
			if record.Address == r.Addr {
				for _, recordTypes := range record.Types {
					if recordTypes.RType == r.Type {
						for _, domain := range recordTypes.Domains {
							if domain == r.Name {
								continue inputLoop
							}
						}
						recordTypes.Domains = append(recordTypes.Domains, r.Name)
						continue inputLoop
					}
				}
				record.Types = append(record.Types, &Type{
					RType:   r.Type,
					Domains: []string{r.Name},
				})
				continue inputLoop
			}
		}
		out.Records = append(out.Records, &Record{
			Address: r.Addr,
			Types: []*Type{
				{
					RType:   r.Type,
					Domains: []string{r.Name},
				},
			},
		})
		continue inputLoop
	}

	return out
}

func (f *FileStore) Sync() error {
	rs, err := f.store.List(context.Background())
	if err != nil {
		return fmt.Errorf("%w: failed to list store records: %v", ErrSync, err)
	}
	b, err := yaml.Marshal(fromEntity(rs...))
	if err != nil {
		return fmt.Errorf("%w: failed to marshal store records to JSON: %v", ErrSync, err)
	}
	err = os.Remove(f.Path)
	if err != nil {
		return fmt.Errorf("%w: failed to remove old reference file: %v", ErrSync, err)
	}
	err = os.WriteFile(f.Path, b, fs.FileMode(store.OS_ALL_RW))
	if err != nil {
		return fmt.Errorf("%w: failed to write new reference file: %v", ErrSync, err)
	}
	return nil
}

func (f *FileStore) Add(ctx context.Context, rs ...*store.Record) error {
	f.mtx.Lock()
	defer func() {
		_ = f.Sync()
	}()
	defer f.mtx.Unlock()

	return f.store.Add(ctx, rs...)
}

func (f *FileStore) List(ctx context.Context) ([]*store.Record, error) {
	return f.store.List(ctx)
}

func (f *FileStore) GetByDomain(ctx context.Context, r *store.Record) (*store.Record, error) {
	return f.store.GetByDomain(ctx, r)
}

func (f *FileStore) GetByDest(ctx context.Context, r *store.Record) ([]*store.Record, error) {
	return f.store.GetByDest(ctx, r)
}

func (f *FileStore) Update(ctx context.Context, domain string, r *store.Record) error {
	f.mtx.Lock()
	defer func() {
		_ = f.Sync()
	}()
	defer f.mtx.Unlock()

	return f.store.Update(ctx, domain, r)
}

func (f *FileStore) Delete(ctx context.Context, r *store.Record) error {
	f.mtx.Lock()
	defer func() {
		_ = f.Sync()
	}()
	defer f.mtx.Unlock()

	return f.store.Delete(ctx, r)
}
