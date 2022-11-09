package memmap

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/zalgonoise/x/dns/store"
)

var (
	test1 = store.New().Name("not.a.dom.ain").Type("A").Addr("192.168.0.10").Build()
	test2 = store.New().Name("really.not.a.dom.ain").Type("A").Addr("192.168.0.15").Build()
	test3 = store.New().Name("really.not.a.dom.ain").Type("CNAME").Addr("am.i.not.a.dom.ain.").Build()
	test4 = store.New().Name("am.i.not.a.dom.ain").Type("A").Addr("192.168.0.15").Build()
)

func TestCreate(t *testing.T) {
	t.Run("OneRecord", func(t *testing.T) {
		ctx := context.Background()
		s := New()

		err := s.Create(ctx, test1)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		records, ok := s.(*MemoryStore).Records[test1.Type]
		if !ok {
			t.Errorf("expected entry for %v to be present in the store, but domain is not assigned", test1)
		}
		addr, ok := records[test1.Name]
		if !ok {
			t.Errorf("expected entry for %v to be present in the store, but record type is not assigned", test1)
		}
		if addr != test1.Addr {
			t.Errorf("stored address %s is incompatible with %v", addr, test1)
		}
	})
	t.Run("ManyRecords", func(t *testing.T) {
		ctx := context.Background()
		s := New()

		err := s.Create(ctx, test1, test2)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		records, ok := s.(*MemoryStore).Records[test1.Type]
		if !ok {
			t.Errorf("expected entry for %v to be present in the store, but domain is not assigned", test1)
		}
		addr, ok := records[test1.Name]
		if !ok {
			t.Errorf("expected entry for %v to be present in the store, but record type is not assigned", test1)
		}
		if addr != test1.Addr {
			t.Errorf("stored address %s is incompatible with %v", addr, test1)
		}

		records, ok = s.(*MemoryStore).Records[test2.Type]
		if !ok {
			t.Errorf("expected entry for %v to be present in the store, but domain is not assigned", test2)
		}
		addr, ok = records[test2.Name]
		if !ok {
			t.Errorf("expected entry for %v to be present in the store, but record type is not assigned", test2)
		}
		if addr != test2.Addr {
			t.Errorf("stored address %s is incompatible with %v", addr, test2)
		}
	})
}

func TestList(t *testing.T) {
	t.Run("EmptyStore", func(t *testing.T) {
		ctx := context.Background()
		s := New()

		rs, err := s.List(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(rs) != 0 {
			t.Errorf("expected store to have zero records; got %v", len(rs))
		}

	})

	t.Run("PopulatedStore", func(t *testing.T) {
		ctx := context.Background()
		s := New()

		err := s.Create(ctx, test1, test2)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		rs, err := s.List(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(rs) != 2 {
			t.Errorf("expected store to have %v records; got %v", 2, len(rs))
		}
	})
}

func TestFilterByDomain(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		s := New()

		err := s.Create(ctx, test1, test2)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		r, err := s.FilterByTypeAndDomain(ctx, test2.Type, test2.Name)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if r == nil {
			t.Errorf("expected a succesful query, returned store.Record is nil")
			return
		}

		if r.Addr != test2.Addr {
			t.Errorf("unexpected output IP address: wanted %s ; got %s", test2.Addr, r.Addr)
		}
	})
	t.Run("FailDoesNotExistDomain", func(t *testing.T) {
		ctx := context.Background()
		s := New()

		err := s.Create(ctx, test1)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		_, err = s.FilterByTypeAndDomain(ctx, test2.Type, test2.Name)
		if !errors.Is(store.ErrDoesNotExist, err) {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("FailDoesNotExistType", func(t *testing.T) {
		ctx := context.Background()
		s := New()

		err := s.Create(ctx, test1, test3)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		_, err = s.FilterByTypeAndDomain(ctx, test2.Type, test2.Name)
		if !errors.Is(store.ErrDoesNotExist, err) {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestFilterByDest(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		s := New()
		wants := []*store.Record{test2, test4}

		err := s.Create(ctx, test1, test2, test3, test4)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		rs, err := s.FilterByDest(ctx, test2.Addr)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(rs) != 2 {
			t.Errorf("unexpected results length: wanted %v ; got %v", 2, len(rs))
		}

		for idx, record := range rs {
			var passes bool
			for _, want := range wants {
				if reflect.DeepEqual(*want, *record) {
					passes = true
					break
				}
			}

			if !passes {
				t.Errorf("output mismatch error: record #%v doesn't match any of [%v %v]: record: %v", idx, wants[0], wants[1], *record)
			}
		}
	})
}

func TestUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		s := New()
		updated := store.New().Name(test1.Name).Type(test1.Type).Addr("192.168.0.9").Build()

		err := s.Create(ctx, test1)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		err = s.Update(ctx, updated.Name, updated)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		records, ok := s.(*MemoryStore).Records[updated.Type]
		if !ok {
			t.Errorf("expected entry for %v to be present in the store, but domain is not assigned", updated)
		}
		addr, ok := records[updated.Name]
		if !ok {
			t.Errorf("expected entry for %v to be present in the store, but record type is not assigned", updated)
		}
		if addr != updated.Addr {
			t.Errorf("stored address %s is incompatible with %v", addr, updated)
		}
	})

	t.Run("FailDoesNotExistDomain", func(t *testing.T) {
		ctx := context.Background()
		s := New()

		err := s.Create(ctx, test1)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		query := store.New().Name(test2.Name).Type(test2.Type).Build()

		err = s.Update(ctx, query.Name, query)
		if !errors.Is(store.ErrDoesNotExist, err) {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("FailDoesNotExistType", func(t *testing.T) {
		ctx := context.Background()
		s := New()

		err := s.Create(ctx, test1, test3)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		query := store.New().Name(test2.Name).Type(test2.Type).Build()

		err = s.Update(ctx, query.Name, query)
		if !errors.Is(store.ErrDoesNotExist, err) {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestDelete(t *testing.T) {
	t.Run("SuccessByDomain", func(t *testing.T) {
		ctx := context.Background()
		s := New()

		err := s.Create(ctx, test1, test2, test3, test4)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		query := store.New().Name(test2.Name).Build()

		err = s.Delete(ctx, query)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		rs, err := s.List(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(rs) != 2 {
			t.Errorf("unexpected output length: wanted %v records ; got %v", 2, len(rs))
		}
	})
	t.Run("SuccessByDomainAndType", func(t *testing.T) {
		ctx := context.Background()
		s := New()

		err := s.Create(ctx, test1, test2, test3, test4)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		query := store.New().Name(test2.Name).Type(test2.Type).Build()

		err = s.Delete(ctx, query)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		rs, err := s.List(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(rs) != 3 {
			t.Errorf("unexpected output length: wanted %v records ; got %v", 3, len(rs))
		}
	})
	t.Run("SuccessByAddress", func(t *testing.T) {
		ctx := context.Background()
		s := New()

		err := s.Create(ctx, test1, test2, test3, test4)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		query := store.New().Addr(test2.Addr).Build()

		err = s.Delete(ctx, query)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		rs, err := s.List(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(rs) != 2 {
			t.Errorf("unexpected output length: wanted %v records ; got %v", 2, len(rs))
		}
	})
}
