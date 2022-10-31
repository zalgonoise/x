package e2e

import (
	"context"
	"reflect"
	"testing"

	"github.com/zalgonoise/x/dns/cmd/config"
	"github.com/zalgonoise/x/dns/dns/core"
	"github.com/zalgonoise/x/dns/health/simplehealth"
	"github.com/zalgonoise/x/dns/service"
	"github.com/zalgonoise/x/dns/store"
	"github.com/zalgonoise/x/dns/store/memmap"
)

var (
	record1 *store.Record = store.New().Type("A").Name("not.a.dom.ain").Addr("192.168.0.10").Build()
	record2 *store.Record = store.New().Type("A").Name("also.not.a.dom.ain").Addr("192.168.0.15").Build()
	record3 *store.Record = store.New().Type("A").Name("really.not.a.dom.ain").Addr("192.168.0.10").Build()
)

func TestService(t *testing.T) {
	s := service.New(
		core.New(),
		memmap.New(),
		simplehealth.New(),
		config.Default(),
	)

	t.Run("Store", func(t *testing.T) {
		t.Run("AddRecords", func(t *testing.T) {
			t.Run("Success", func(t *testing.T) {
				ctx := context.Background()

				err := s.AddRecords(ctx, record1, record2)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
			})
		})

		t.Run("ListRecords", func(t *testing.T) {
			t.Run("Success", func(t *testing.T) {
				ctx := context.Background()

				rs, err := s.ListRecords(ctx)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if len(rs) != 2 {
					t.Errorf("unexpected returned records length: wanted %v ; got %v", 2, len(rs))
					return
				}
				pass := [2]bool{}
				for idx, r := range rs {
					if !reflect.DeepEqual(record1, r) && !reflect.DeepEqual(record2, r) {
						t.Errorf("output record %v does not match the expected", r)
						continue
					}
					pass[idx] = true
				}
				for _, ok := range pass {
					if !ok {
						t.Errorf("expected both entries to be returned")
						return
					}
				}
			})
		})

		t.Run("GetRecordByTypeAndDomain", func(t *testing.T) {
			t.Run("Success", func(t *testing.T) {
				ctx := context.Background()

				r, err := s.GetRecordByTypeAndDomain(ctx, record1.Type, record1.Name)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if !reflect.DeepEqual(record1, r) {
					t.Errorf("output mismatch error: wanted %v ; got %v", record1, r)
					return
				}
			})
		})

		t.Run("GetRecordByAddress", func(t *testing.T) {
			t.Run("Success", func(t *testing.T) {
				ctx := context.Background()

				rs, err := s.GetRecordByAddress(ctx, record2.Addr)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if len(rs) != 1 {
					t.Errorf("unexpected returned records length: wanted %v ; got %v", 1, len(rs))
					return
				}
				if !reflect.DeepEqual(record2, rs[0]) {
					t.Errorf("output mismatch error: wanted %v ; got %v", record2, rs[0])
					return
				}
			})
		})

		t.Run("UpdateRecord", func(t *testing.T) {
			t.Run("Success", func(t *testing.T) {
				ctx := context.Background()

				err := s.UpdateRecord(ctx, record1.Name, record3)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				r, err := s.GetRecordByTypeAndDomain(ctx, record3.Type, record3.Name)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if !reflect.DeepEqual(record3, r) {
					t.Errorf("output mismatch error: wanted %v ; got %v", record3, r)
					return
				}
			})
		})

		t.Run("DeleteRecord", func(t *testing.T) {
			t.Run("SuccessByTypeAndDomain", func(t *testing.T) {
				ctx := context.Background()
				toDelete := store.New().Type(record3.Type).Name(record3.Name).Build()

				err := s.DeleteRecord(ctx, toDelete)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				rs, err := s.ListRecords(ctx)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if len(rs) != 1 {
					t.Errorf("unexpected returned records length: wanted %v ; got %v", 1, len(rs))
					return
				}
				if !reflect.DeepEqual(record2, rs[0]) {
					t.Errorf("output mismatch error: wanted %v ; got %v", record2, rs[0])
					return
				}
			})
		})
	})
}
