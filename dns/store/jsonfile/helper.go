package jsonfile

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"

	"github.com/zalgonoise/x/dns/store"
)

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

func (f *FileStore) sync() error {
	rs, err := f.store.List(context.Background())
	if err != nil {
		return fmt.Errorf("%w: failed to list store records: %v", store.ErrSync, err)
	}
	b, err := json.Marshal(fromEntity(rs...))
	if err != nil {
		return fmt.Errorf("%w: failed to marshal store records to JSON: %v", store.ErrSync, err)
	}
	err = os.WriteFile(f.Path, b, fs.FileMode(store.OS_ALL_RW))
	if err != nil {
		return fmt.Errorf("%w: failed to write new reference file: %v", store.ErrSync, err)
	}
	return nil
}
